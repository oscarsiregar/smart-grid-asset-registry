# Research: Asset Explorer

**Feature**: [spec.md](spec.md) · **Plan**: [plan.md](plan.md)

Decisions taken while planning, with the reasons and the alternatives. Feature 001's decisions
([../001-csv-import/research.md](../001-csv-import/research.md)) still apply; this file only adds what the
explorer needs. Reference modules: backend `asset_import` and `domain/assetimport`, frontend `modules/imports`
and `pages/assets`.

## R0. Template fit: deviations that need a decision

Checked against the constitution (v1.1.0), `backend/AGENTS.md` + `backend/docs/rules/*`, `frontend/AGENTS.md` +
`frontend/docs/code-conventions.md`.

| # | Template rule | What this feature needs | Status |
|---|---|---|---|
| **D1** | List endpoints are paged: `GetListXxxInput.SetDefault()` fills a default `limit` and caps it (`04-usecase.md`); "the usecase always fills the default" (`05-repository-database.md`, *Paging list*) | The tree's top level and a node's children are returned **complete, unpaged** (clarification Q1: show all children at once, no paging; FR-003) | **Needs a new Template Exception (13)** — not worked around |

Everything else fits the template without an exception:

- New backend module `asset` copies the `asset_import` shape (handler / usecase / repository, one file per
  endpoint, use case and repository method); the existing `psql/asset` repository gains read methods.
- Routes stay public under `/api` (exceptions 1, 5); responses use the go-common envelope and the list shape
  `{ items, totalItems }` (`03-handler.md` rule 7).
- The not-found case is a business error through `internal/apperr` with a new code and an English message
  (exception 6).
- No schema change: the `asset` table and its indexes from feature 001 already serve every read (R4).
- Frontend: a new `modules/assets` and new components under `pages/assets`, no new dependency.

**Proposed Exception 13 — Complete hierarchy lists.** "The asset hierarchy reads (top-level substations, the
children of one asset) return every row, without `page`/`limit`. The hierarchy is browsed as a whole (spec
002, clarification Q1), and the 5 MB import limit bounds the registry to the order of 50 000 assets. Other list
endpoints keep the template's paging." Adding it is a constitution amendment (MINOR, 1.1.0 → 1.2.0) through
`/speckit-constitution` before `/speckit-implement`.

Alternatives to D1, rejected:
- *Paged endpoints, frontend fetches every page*: works around the rule instead of following it, and adds a loop
  of requests per node.
- *Paging with a very high default limit (e.g. 100 000)*: same behaviour as no paging, but hides the decision
  behind a magic number.
- *"Show more" in the UI*: contradicts clarification Q1.

## R1. API shape

**Decision**: Two read endpoints in a new `asset` module:

- `GET /api/assets` — the top level (every substation); `GET /api/assets?parentAssetId=<id>` — the immediate
  children of one asset. One usecase `GetListAsset`, list shape `{ items, totalItems }`, items are *asset
  nodes* (ID, name, type, status, child count).
- `GET /api/assets/:assetId` — one asset's details for the details area: all attributes, its path from the
  substation down to itself, its immediate children, and, for a substation, the counts of each type beneath it.
  Usecase `GetAssetDetail`.

Details are in [contracts/assets-api.md](contracts/assets-api.md).

**Rationale**: Follows the template's REST table (`GET /<plural>` list, `GET /<plural>/:id` detail) and method
names (`GetListXxx`, `GetXxxDetail`). The tree loads a node's children only when it is expanded (user input),
one request per expanded node. The detail answer carries everything the details area shows, so selecting an
asset is one request; its `path` is also what the tree needs to expand down to a deep-linked asset.

**Alternatives**: one endpoint returning the whole tree (simple, but loads 50 000 assets to show 13
substations, and contradicts "load children when expanded"); a separate `/children` sub-resource (a second list
endpoint with the same item shape — the `parentAssetId` filter is the template's list-filter pattern);
returning ancestors' siblings in the detail (saves 1–2 requests on a deep link, but couples the detail to the
tree's layout).

## R2. Asset IDs in addresses

**Decision**: The API and the browser use the natural `asset_id`, percent-encoded in paths.

- **Backend**: Fiber 2.52 keeps a path parameter **encoded** (probe: `/api/assets/SS%2F1` matches `/:assetId`
  with the value `SS%2F1`; `%20`, `%25`, `%3F`, `%23` likewise). The handler decodes it with
  `url.PathUnescape`; an invalid escape is `400 INVALID_REQUEST`. An encoded `/` never splits the route.
  `parentAssetId` is a query parameter, which Fiber already decodes.
- **Proxies**: nginx `proxy_pass http://backend;` (no URI part) and the Vite dev proxy forward the request
  path unchanged, so `%2F` reaches the API encoded. nginx's SPA fallback (`try_files $uri /index.html`) serves
  `index.html` for any `/assets/...` path.
- **Frontend**: React Router 7.18 (probe): `/assets/:assetId` matches `/assets/SS%2F1` with the param `SS/1`,
  and decodes `%20`, `%25`, `%3F`, `%23`; `generatePath` encodes all of them. Links are built with
  `generatePath(AppRoute.ASSETS_$ASSET_ID, { assetId })`.

**Rationale**: FR-017 requires every stored ID to round-trip; feature 001 accepts any trimmed, non-blank text
as an ID. The only exceptions are `.` and `..`: the URL standard treats them (encoded or not) as dot segments,
and both the browser address and `fetch` rewrite the path (`/api/assets/%2E%2E` becomes `/api/`), so neither
their address nor their detail request can be built. They stay listed and expandable in the tree (the
children request carries the ID as a query parameter) but are not selectable (spec edge case,
KNOWN_LIMITATIONS; decision 2026-11-19). The natural key keeps addresses readable (`/assets/SS-001`) and matches the `asset_id` foreign key
(001 R7). Both probes are turned into tests (handler test, router test).

**Alternatives**: UUID in the address (stable, but unreadable and not what the brief shows); ID in a query
parameter (`/assets?asset=SS-001`) (avoids path encoding, but the brief asks for `/assets/<asset_id>`).

## R3. Ordering

**Decision**: The backend orders every list (FR-004) with one pure function in `domain/entity`:
`SortAssetNodes` — by type rank (SUBSTATION, TRANSFORMER, LV_BOARD, SWITCHBOARD, SWITCHBOARD_PANEL, from a new
`valueobject.AssetType.Rank()`), then by asset ID compared *naturally*: the ID is split into runs of digits and
non-digits; digit runs compare as numbers (leading zeros ignored, then shorter-first as tie-breaker),
other runs compare by Unicode code point, and where a digit run meets a non-digit run the digit run comes first
(so in the sample `SS-001` … `SS-012` come before `SS-T01`). Equal keys fall back to plain string comparison, so the order is total
and deterministic. The frontend shows lists in the order received and never re-sorts.

**Rationale**: Ordering is a rule of the feature (clarification Q4), so it lives in the backend (principle II)
and is unit-tested table-driven (principle IV). `domain/entity` already holds pure asset logic
(`DifferingFields`) and its tests run in `make check`, so no new package or exception is needed. Sorting in Go
over a complete list is cheap (50 000 IDs sort in milliseconds).

**Alternatives**: `ORDER BY asset_id` (plain text: `SS-10` before `SS-2`, contradicts Q4); an ICU collation
with numeric ordering (`en-u-kn-true`) (a migration plus a dependency on the database image's ICU build, and
harder to test without a database); sorting in the frontend (business rule outside the backend).

## R4. Reading the hierarchy

**Decision**: New read methods on the existing `AssetRepository` (one file each, `05-repository-database.md`):

| Method | Query | Notes |
|---|---|---|
| `FindByAssetID(assetID)` | `WHERE asset_id = ?` | `nil, nil` when not found |
| `FindNodes(filter{ParentAssetID *string})` | `parent_asset_id IS NULL` (top level) or `= ?`, with `(SELECT count(*) FROM asset c WHERE c.parent_asset_id = a.asset_id) AS child_count` | returns `[]*entity.AssetNode`, unsorted |
| `FindPath(assetID)` | recursive CTE upwards from the asset (`asset` joined to its parent) | the asset and its ancestors, substation first; empty when not found |
| `CountDescendantsByType(assetID)` | recursive CTE downwards from the asset, `GROUP BY asset_type` | `map[AssetType]int`, types without descendants absent |

All SQL uses `?` placeholders (`db.Raw` for the two CTEs). No migration: `ux_asset_asset_id` serves lookups and
the path walk, `ix_asset_parent_asset_id` serves children, child counts and the downward walk. The top-level
query (`parent_asset_id IS NULL`) scans the table; at 50 000 rows that is a few milliseconds, so no partial
index is added now (measured in the integration test, R10).

**Rationale**: The recursive CTEs do not depend on the hierarchy being three levels deep, so a future type
cannot silently break the counts; they are one round trip each. Counting stays a data-access concern; which
four types are shown, and only for substations, is the usecase's decision.

**Alternatives**: loading the whole subtree and counting in Go (transfers every descendant to count them);
fixed two-level joins (correct today, wrong as soon as the depth changes); a materialised path column
(schema change and import change for no current need).

## R5. Detail use case and consistency

**Decision**: `GetAssetDetail` reads, without a transaction: `FindByAssetID` (nil → `404 ASSET_NOT_FOUND`), then
`FindPath`, `FindNodes(parent = the asset)`, and — only for a substation — `CountDescendantsByType`, mapped to
the four counts (transformers, LV boards, switchboards, switchboard panels; absent types are 0). Lists are
sorted with `SortAssetNodes`. `GetListAsset` with an unknown `parentAssetId` is also `404 ASSET_NOT_FOUND`
(checked with `FindByAssetID` when the list is empty).

**Rationale**: The template uses transactions for writes and read-then-write (`04-usecase.md`); these are pure
reads. Assets are only ever added, in one committed transaction per import (001 R5), so the reads can only
straddle an import commit: the counts could then include new assets the children list does not yet show, until
the next refetch, which the frontend does right after a committed import (R7). Recorded in
`docs/KNOWN_LIMITATIONS.md`.

**Alternatives**: a `REPEATABLE READ` read-only transaction (exact snapshot, but `gormtx` has no isolation
option today and the window is tiny); one large SQL statement returning everything (hard to read and test).

## R6. Not found and other errors

**Decision**: New code `ASSET_NOT_FOUND` in `domain/errorcode` with the English message "No asset with ID %s
exists." (`404`). Bad percent-encoding in the path is `400 INVALID_REQUEST`. Technical errors are `500` as
usual. The frontend treats `ASSET_NOT_FOUND` on the detail request as the *not found* state (FR-020), not as an
error, and suppresses the toast for both asset endpoints because the page renders every failure itself.

**Rationale**: The template's usecase returns `apperr.NotFound` when a repository returns `nil, nil`; the
specific code lets the frontend tell "no such asset" from "server failed" without parsing text.

**Alternatives**: `200` with `null` data (hides the case from the template's error handling); a generic
`NOT_FOUND` derived from the status (also produced by unknown routes and proxies, so ambiguous).

## R7. Frontend data and caching

**Decision**: New `modules/assets` with two RTK Query **queries** and a tag:

- `getAssetNodes(parentAssetId | null)` → node list; `getAssetDetail(assetId)` → detail. Both
  `providesTags: [AssetsApiTag.ASSET]`, `extraOptions: { shouldSuppressToast: true }`.
- The tree calls `getAssetNodes(null)` for the top level and `getAssetNodes(id)` inside a node **only while it
  is expanded** (the query hook mounts with the child list), so children load on first expansion and come from
  the cache afterwards.
- `createImport` in `modules/imports` gets `invalidatesTags: (result) => result?.isCommitted ? [ASSET] : []`, so
  after a committed import every mounted asset query refetches (FR-026), and nothing refetches otherwise.
  `imports-api` registers the tag type with `enhanceEndpoints({ addTagTypes: [AssetsApiTag.ASSET] })` and imports
  the enum from the `modules/assets` barrel.

**Rationale**: Matches the conventions (§8: queries `providesTags`, mutations `invalidatesTags` exactly what
they change) and the user's "load a node's children only when it is expanded". One coarse tag is right because
an import can add assets anywhere in the tree.

**Alternatives**: per-node tags (`{ type: ASSET, id }`) — an import does not report which parents it touched;
manual `refetch` calls from the import panel (couples the panel to the tree).

## R8. Tree component

**Decision**: A hand-written, page-specific tree in `pages/assets/components/asset-tree`, following the WAI-ARIA
*navigation treeview* pattern:

- `role="tree"`, nodes `role="treeitem"` with `aria-level`, `aria-expanded` (only on nodes with children),
  `aria-selected` on the selected asset, children in `role="group"`.
- Each node's label is a link (`<Link to={assetPath}>`), so selecting is navigating: the address changes, a
  history entry is added (FR-018), and open-in-new-tab works. A separate toggle expands and collapses.
- Keyboard (FR-007): roving `tabIndex`; Up/Down move between visible nodes; Right expands or moves to the first
  child; Left collapses or moves to the parent; Home/End; Enter selects. Implemented in one page hook
  (`use-tree-keyboard`).
- Expansion state is local (`useState` of expanded IDs). When the selected asset changes (address or link) and
  its path is known, its ancestors are added to the expanded set during render (the React "adjust state when a
  prop changes" pattern, comparing the previous selection), so the user can still collapse them afterwards
  (spec edge case). No effect copies query data into state.
- The selected node is marked with `aria-selected`, a visible indicator and bold text (not colour alone,
  FR-006) and scrolled into view by a small effect on its element when it is selected through the address or a
  link.
- Status label only for assets not in service (clarification Q3), in words.

**Rationale**: No tree component exists in the template and none is worth a dependency (the conventions ask for
a reason the current stack cannot cover). The navigation-tree pattern gives correct semantics for "selecting
is opening a page". Per-node lazy children fall out naturally from mounting the child list only when expanded.

**Alternatives**: a tree library (new dependency, styling and ARIA to override); nested `<details>` elements (no
tree semantics or arrow-key navigation); buttons instead of links (no history entry or new-tab).

## R9. Routes and page layout

**Decision**: `AppRoute.ASSETS_$ASSET_ID = "/assets/:assetId"` next to `AppRoute.ASSETS`, both rendering the
`Assets` page (two lazy route entries, same chunk). The page reads `assetId` from `useParams`; `/assets` (and
`/assets/` with a blank segment) means no selection (FR-021). Layout: page intro and **Import CSV** button as
today, then a two-column explorer — tree on the left, details area on the right, stacked on narrow screens. The
import panel keeps `?panel=import` (`usePanel` changes only search parameters, so the asset path is kept,
FR-025). With no stored assets the explorer is replaced by the empty state with an **Import CSV** button.

**Rationale**: Reuses the existing page and panel (user input), and keeps the selection in the path as the
brief and principle V require.

**Alternatives**: a nested route with an `<Outlet>` for the details (more structure for one page); the
selection in a search parameter (not the brief's address).

## R10. Testing

**Decision**:
- **Backend unit**: `SortAssetNodes` and `AssetType.Rank` table tests (written first); usecase tests with mocks
  for every branch (list top level / children / unknown parent; detail found for each type, not found, counts
  only for substations, absent types → 0, repository errors); handler tests (`app.Test`) for status codes, the
  envelope, `%2F` / `%20` / `%25` decoding and a bad escape; repository tests with sqlmock.
- **Backend integration** (`internal/integrationtest`, real PostgreSQL): import the sample through the feature
  001 usecase, then: 13 top-level substations; every one of the 205 assets reachable by expanding from the top
  (SC-001); for each substation, `CountDescendantsByType` equals the count obtained by walking `FindNodes`
  (SC-002); `FindPath` for a panel is substation → switchboard → panel; an ID with `/` and spaces round-trips.
  With the 50 000-row file from 001's large-file test (thousands of substations, so it stresses the top
  level) plus one switchboard with 5 000 panels (a wide node): the top level and that node's children each
  within the SC-006 budget.
- **Frontend** (Vitest + RTL, `fetch` mocked): tree renders the top level and requests children only on
  expansion; keyboard navigation; deep link expands ancestors and marks the asset selected; not-found state;
  empty state; error with retry per area; details fields with "Not recorded"; substation groups and counts;
  switchboard panels; path links; a committed import refetches, a non-committed one does not.

## R11. Performance (SC-006)

**Decision**: No virtualisation and no caching beyond RTK Query. Budgets checked in the integration test (API
side) and by hand in the browser with the 50 000-row file: top level ≤ 2 s, expand / select ≤ 1 s.

**Rationale**: Clarification Q1 chose complete lists; with the sample a node has at most a few dozen children,
and even a node with thousands of children renders a few thousand list items, which React handles within the
budget. Correctness and readability come before optimisation (brief).

**Alternatives**: windowed rendering (a dependency or a lot of code; deferred until a real registry needs it,
recorded in `docs/KNOWN_LIMITATIONS.md`).
