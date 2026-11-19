# Architecture

How the Smart Grid Asset Registry is put together, and why. It covers two features:
[specs/001-csv-import](../specs/001-csv-import/spec.md) — import a grid-asset CSV, validate it on the server,
store the accepted rows — and [specs/002-asset-explorer](../specs/002-asset-explorer/spec.md) — browse the
stored hierarchy. The decisions behind each part are recorded in each feature's research
([001](../specs/001-csv-import/research.md) R0–R14, [002](../specs/002-asset-explorer/research.md) R0–R11; cited
below as R*n* and 002-R*n*); the project rules are in the [constitution](../.specify/memory/constitution.md).

## 1. System overview

```
browser ──► web (nginx, :3000) ──/api──► api (Go/Fiber, :8080) ──► db (PostgreSQL 16, :55432)
             serves the SPA                validates + stores          asset_registry
                                                                       asset_registry_test (tests)
```

`docker-compose.yml` runs three services (Template Exception 3):

| Service | What it is | Notes |
|---|---|---|
| `db` | PostgreSQL 16 | Host port 55432. An init script also creates `asset_registry_test` for integration tests. |
| `api` | The Go backend | Applies the embedded migrations on start, then serves `/api`. Healthy when `/api/health` answers. |
| `web` | nginx with the built React app | Proxies `/api` to `api:8080`; every other path falls back to `index.html`. |

The `test` profile adds `backend-test` (`make check` with `TEST_DATABASE_URL` pointing at `db`) and
`frontend-test` (`pnpm test`), so the Definition of Done checks run without local toolchains.

There is no authentication (Template Exception 1): every route is public under `/api`.

## 2. Backend

### Layers

The backend keeps the company template's clean architecture (`backend/AGENTS.md`):

```
handler/asset_import  ──►  usecase/asset_import  ──►  repository/psql/{asset, asset_import}
 (multipart, DTOs)          (flow, transaction)         (GORM, one file per method)
handler/asset         ──►  usecase/asset         ──►  repository/psql/asset (read methods)
 (query/path, DTOs)         (reads, ordering, counts)
        └──────────── all depend only on interfaces in domain/ ────────────┘
                                   │
            domain/assetimport (pure import rules) · domain/entity (asset ordering)
```

- **Handler** (`internal/handler/asset_import`): reads the `file` and `mode` multipart parts, calls the usecase,
  and maps the report to the response DTO. Row issues carry codes; the handler renders their English text
  through the `internal/apperr` catalog (Template Exception 6).
- **Usecase** (`internal/usecase/asset_import`): parses the file, loads the stored assets it mentions, runs the
  rules, and, for a commit, writes inside one transaction. It sets IDs, timestamps and the audit name
  `"system"`.
- **Repositories**: `asset` (`FindByAssetIDs`, `FindBySerialNumbers`, `CreateMany`) and `asset_import`
  (`Create`, `AcquireLock`). Lookups send one array parameter (`= ANY(?)`), so a 50 000-row file never reaches
  PostgreSQL's bind-parameter limit; inserts are batched in 1 000s (R6).
- **go-common**: the company's private library is re-implemented in `backend/pkg/gocommon` with the same names
  and signatures (Template Exception 4). Its only addition is `data.errorCode` on error responses.

### The import engine (`domain/assetimport`)

Every import rule lives in one pure package that imports only the standard library and other `domain`
packages (Template Exception 10, R2). It takes the parsed file and a map of the stored assets the file
mentions, and returns a `Report`: accepted assets in insert order, unchanged rows, rejected rows with every
reason, warnings and counts. Issues are codes with arguments, never text. Because it does no I/O, every rule
is tested without mocks, and the sample file is pinned by a hand-derived golden test (`testdata/`).

`Parse` refuses a whole file for file-level problems: over 5 MB, no data rows, not UTF-8, any CSV syntax
error, missing or repeated columns. Those become `422` responses with a specific `data.errorCode`, and no row
results.

### Order-independent validation (R3)

`Validate` never walks the rows in sequence. Every decision uses maps built from the whole file, so any
shuffle of the same rows gives the same report (FR-017):

1. **Field pass**, per row: required fields, asset type, status, numbers, dates, plus duplicate asset IDs
   counted over the whole file. A row with the wrong number of fields reports only `MALFORMED_ROW`.
2. **Hierarchy pass**, only on rows whose fields are valid:
   - parent presence (substations have none, everything else needs one), self-parent, parent not found in the
     file or the store;
   - the parent's type must be the permitted parent type (in-file row or stored asset);
   - cycles, found by following parent pointers over the rows with a unique ID (every row on a cycle is
     rejected).
3. **Re-import pass**: a row whose ID is already stored is *unchanged* when every CSV column matches
   (numbers by value, dates by calendar day), otherwise rejected with `ASSET_ID_CONFLICT` naming the
   differing columns. Stored assets are never edited.
4. **`PARENT_REJECTED`**, repeated until nothing changes: a row whose in-file parent is rejected is rejected
   too, so the whole subtree below a bad row goes, including the children of a conflicting row.
5. **Serial-number warnings** on the rows that are not rejected: a serial shared with another such row or a
   stored asset raises `DUPLICATE_SERIAL_NUMBER`, which never blocks the row.

Accepted assets are sorted by depth within the file, then by asset ID, so the batched insert never refers to a
parent that is not stored yet.

### Validate versus commit (R4)

| Endpoint | Does | Stores | Answers |
|---|---|---|---|
| `POST /api/imports/validate` | parse + validate against the current store | nothing | `200`, outcome `VALIDATED_ONLY` |
| `POST /api/imports` | the same, inside a transaction, then writes | an import record (+ assets) | `201` |

The server keeps no state between the two calls: the browser sends the same file again on Save, and the
server re-validates it at that moment (FR-022). A parent stored in between makes a row valid; a conflicting
asset stored in between makes it rejected.

### Transaction, lock and outcomes (R5)

`CreateAssetImport` runs in one transaction:

1. `SELECT pg_advisory_xact_lock(<fixed key>)`: imports are serialised, so two imports never decide against
   the same stale store. Readers of `asset` are not blocked. The lock is released at commit or rollback.
2. Load the stored assets the file mentions, then run `Validate`.
3. Decide the outcome and insert the import record:

| Condition (checked in this order) | Outcome | Assets inserted |
|---|---|---|
| strict mode and at least one rejected row | `STRICT_BLOCKED` | none |
| nothing new to import | `NOTHING_TO_COMMIT` | none |
| otherwise | `COMMITTED` | every accepted row |

4. Commit. Any technical error rolls everything back: no assets, no record, a `500`, and an error log with
   the file name, mode and the step that failed (FR-030).

Only completed attempts store a record (`asset_import`); validate-only runs and failed attempts do not, and
row-level details are never stored (FR-031).

### Schema (R7)

```
asset_import (id PK, file_name, mode, outcome, total/imported/unchanged/rejected/warning_rows, audit columns)
      ▲ import_id (FK)
asset (id PK, asset_id UNIQUE, parent_asset_id FK → asset(asset_id), asset_type, asset_name,
       operational_status, voltage_kv NUMERIC, rating_kva NUMERIC, manufacturer, model,
       serial_number, commissioned_date DATE, audit columns)
```

- The hierarchy is a self-reference on the natural key `asset_id`, the same ID the CSV uses. Parent links stay
  readable and need no ID translation.
- Blank CSV values are stored as `NULL`; free text is `TEXT`, so there is no hidden length rule.
- Numbers are unconstrained `NUMERIC`, so values are stored exactly as typed.
- There are no CHECK constraints (Template Exception 8): types, statuses, parent rules and `rating_kva ≥ 0` are
  the backend's. The unique key and both foreign keys remain as the database's last guarantee.
- The template's audit columns stay; with no users, `created_by_name` and `updated_by_name` are `"system"`.

Migrations are embedded in the binary and applied with golang-migrate when the API starts (R8).

### The asset explorer reads (feature 002)

Two read-only endpoints in the module `asset` ([API reference](../backend/docs/api/asset.md)):
`GET /api/assets[?parentAssetId=]` lists the top level or one asset's children, and
`GET /api/assets/:assetId` returns one asset with its path, its children and — for a substation — the counts of
each type beneath it (002-R1). No table or migration was added; the existing indexes serve every query.

- **Lists are complete, not paged** (Template Exception 13, 002-R0): the tree shows every child of a node at
  once. Only the hierarchy reads are exempt; other list endpoints keep the template's paging.
- **Reads** (002-R4), one repository method each: `FindByAssetID`; `FindNodes` (the top level or one parent's
  children, each with a `child_count` subquery so the tree knows what is expandable); `FindPath`, a recursive
  CTE walking up `parent_asset_id`; and `CountDescendantsByType`, a recursive CTE walking down and grouping by
  type. The CTEs do not assume the hierarchy is three levels deep. Every request value is a bound `?`.
- **Ordering** (002-R3) is a pure function, `entity.SortAssetNodes`: by type rank (`AssetType.Rank()`), then by
  asset ID in natural order (`CompareAssetIDs`: digit runs compare as numbers, `SS-2` before `SS-10`, and a
  digit run comes before a letter, so `SS-012` comes before `SS-T01`). It is total, so equal input always gives
  the same order, and it is table-tested. The frontend shows lists as received.
- **Detail flow** (002-R5): `FindByAssetID` (missing → `404 ASSET_NOT_FOUND`), `FindPath`, the children, and
  the counts only for a substation; which four types are counted is the usecase's decision. The reads run
  without a transaction: assets are only ever added, so at most they straddle an import commit.
- **Addresses** (002-R2): the natural `asset_id` is the address, percent-encoded. Fiber keeps path parameters
  encoded, so the handler decodes `:assetId` (`SS%2F9%20A` → `SS/9 A`; a malformed escape is `400`).
  `parentAssetId` is a query value and arrives decoded.
- **Scale** (002-R11): with about 55 000 stored assets the integration test measures the top level, a
  5 000-child list and a substation's details in well under 100 ms each.

## 3. Frontend

React 19 + TypeScript, RTK Query, React Hook Form + Zod, CSS Modules, on the company frontend template
(`frontend/AGENTS.md`). The structure follows R10:

- `/` redirects to `/assets`, the main page: the asset explorer, or an empty state while nothing is stored.
- **Import CSV** opens the import panel over the page. The panel is part of the URL (`?panel=import`), so
  refresh, Back and shared links keep it. It is a native `<dialog>` opened with `showModal()`: the browser traps
  focus and handles Escape, and focus returns to the button that opened it.
- `modules/imports` holds the two mutations (`validateImport`, `createImport`) sending `FormData`, the DTO types
  written from the API contract, and a schema that only checks a file is chosen. Every rule about the file's
  content is the server's (constitution, principle II).
- The panel shows results straight from the mutation state, never a copy. After **Check file** it shows
  "Validated only — nothing saved", the counts, every rejected row with every reason, and the warnings. **Save N
  assets** then resends the same file and mode. Every result states in words whether data was saved; colour is
  never the only signal (principle V).
- `baseQuery` unwraps the go-common success envelope. On errors it takes the code from `data.errorCode`, or
  derives one from the HTTP status (`frontend/docs/code-conventions.md` §11).

The asset explorer (feature 002) lives in the same page:

- **Address as state** (002-R2, 002-R9): the selected asset is the path, `/assets/:assetId`. Selecting an asset
  is following a link, so every selection is a history entry, Back and Forward walk through them, and a
  refresh or a shared link reopens it. Links are built with `getAssetPath`, which percent-encodes the ID through
  `generateRoutePath`; the shared `Link` only accepts such generated addresses or fixed routes.
- **Data** (002-R7): `modules/assets` has two queries, `getAssetNodes(parentAssetId | null)` and
  `getAssetDetail(assetId)`, both providing one `Asset` cache tag. `createImport` invalidates that tag only when
  the result says data was committed, so after a commit every mounted list and the details refetch, and an
  import that saved nothing changes nothing.
- **Tree** (002-R8): a hand-written WAI-ARIA navigation tree. A node's children are a separate component with its
  own query, mounted only while the node is expanded, so children load on first expansion and come from the
  cache afterwards. `useTreeKeyboard` gives a roving tab stop and the arrow, Home, End and Enter keys.
  `useExpandedNodes` keeps the expanded set and, when the selected asset's path arrives, expands its ancestors
  once per selection, adjusting state during render rather than copying query data in an effect. The selected
  node is marked in words ("Selected") and bold, not by colour alone, and scrolled into view.
- **Details**: the path (ancestors are links), every attribute (blanks say "Not recorded"), for a substation its
  children in three groups and the four counts, for a switchboard its panels. `ASSET_NOT_FOUND` is a
  "not found" state with a way back, not an error and not a toast.

## 4. Key trade-offs

| Decision | Chosen | Instead of | Why |
|---|---|---|---|
| Where rules live (R2) | pure `domain/assetimport` | usecase helpers, SQL | Testable without mocks; one source of truth for both endpoints |
| Rule evaluation (R3) | passes over maps, fixpoint propagation | single ordered pass | Row order can never change the result |
| Confirm flow (R4) | resend the file, re-validate | server-side preview token | Stateless; the store may change between check and save |
| Concurrency (R5) | transaction-scoped advisory lock | `SERIALIZABLE`, table lock | No retries, readers not blocked; constraints stay as a backstop |
| Stored-asset loading (R6) | lookups of only the IDs and serials in the file | read the whole table | Scales with the file, not the registry |
| Parent key (R7) | FK on natural `asset_id` | FK on the UUID | Readable links, no translation on import |
| Import panel (R10) | `<dialog>` + URL parameter | separate route, dialog library | Correct focus handling, no new dependency, survives refresh |
| Hierarchy API (002-R1) | top-level/children list + one detail endpoint | whole tree in one response, `/children` sub-resource | Loads only what is expanded; the template's list-filter pattern |
| Complete lists (002-R0) | no paging for hierarchy reads (Exception 13) | paging, "show more", frontend looping over pages | The tree shows every child at once; one request per node |
| Addresses (002-R2) | natural `asset_id`, percent-encoded in the path | UUID in the address, ID as a query parameter | Readable addresses as the brief shows; IDs `.` and `..` are the one exception |
| Ordering (002-R3) | natural order in a pure Go function | `ORDER BY asset_id`, ICU numeric collation | `SS-2` before `SS-10`; testable without a database, no migration |
| Hierarchy reads (002-R4) | recursive CTEs for path and counts | loading the subtree into Go, fixed joins | One round trip each; correct at any depth |
| Tree (002-R8) | hand-written navigation tree | a tree library | Full keyboard and ARIA control, no new dependency |

Known limits and simplifications are listed in [KNOWN_LIMITATIONS.md](KNOWN_LIMITATIONS.md).
