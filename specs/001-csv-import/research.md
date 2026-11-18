# Research: CSV Import of Grid Assets

**Feature**: [spec.md](spec.md) · **Plan**: [plan.md](plan.md) · **Date**: 2026-11-19

Each entry: **Decision**, **Rationale**, **Alternatives considered**. Section R0 lists the places where
the template rules do not fit the spec or constitution; nothing there is silently worked around.

---

## R0. Template fit: deviations that need a decision

Principle I says every deviation is listed under *Template Exceptions* **before** it is made. The
items below were not covered by exceptions 1–7 and were proposed as a constitution amendment.

**Resolution (constitution 1.1.0, 2026-11-19)**: D1, D3, D4, D5 and D6 were accepted as exceptions
8, 9, 10, 11 and 12. **D2 was rejected**: both tables keep the template's `created_by_name` /
`updated_by_name` audit columns, filled with the fixed value `"system"` (see R7). The table below
keeps the original proposals for the record.

| # | Template rule (source) | Conflict | Proposal → outcome |
|---|---|---|---|
| D1 | Enums stored `VARCHAR + CHECK (status IN (...))`; `price CHECK (price >= 0)` (backend `05-repository-database.md`, product migration) | Constitution II forbids CHECK constraints | No CHECK constraints; enums are plain `VARCHAR` validated by the backend → **accepted, exception 8** |
| D2 | Every table has `created_by_name` / `updated_by_name` audit columns (`05`) | No users (exception 1), so there is no actor name | Omit the `*_by_name` columns → **rejected**: keep the columns, value `"system"` |
| D3 | Docs (`docs/`, README) in Indonesian; catalog entries carry `ID` + `EN` (`02-penamaan.md`, `06-error-handling.md`) | Exception 6 covers only the message catalog; it is silent on docs. The brief's reviewers read English | All project documentation (root `docs/`, `backend/docs/api/*.md`, READMEs) in English; existing Indonesian rule files stay → **accepted, exception 9** |
| D4 | Folder layout has no place for pure domain logic beyond `Validate()` and entity methods (`01-arsitektur.md`) | The import rules are a large, pure, exhaustively tested engine (user decision: no DB/HTTP/framework imports) | New package `backend/domain/assetimport/` (stdlib + other `domain/*` only), one file per concern → **accepted, exception 10** |
| D5 | Integration tests: template has only sqlmock repository tests and mocked usecase tests (`08-testing.md`); usecase packages must not import `internal/repository` (`01`) | Constitution IV requires real-PostgreSQL tests of repositories and usecases | Test-only packages `backend/internal/integrationtest/` and `backend/internal/testsupport/pgtest/`, skipped when `TEST_DATABASE_URL` is unset → **accepted, exception 11** |
| D6 | File uploads: "validate MIME from content, server-generated name, object storage, presigned URL" (`07-keamanan.md`); handlers parse with `BodyParser` only (`03-handler.md`) | The CSV is parsed in memory and never stored; it arrives as `multipart/form-data` | Handler reads the part with `ctx.FormFile("file")`; the bytes go to the usecase as an `io.Reader`; no storage → **accepted, exception 12** |

Items that were already covered: no auth / no permission checks (exception 1),
Vitest + RTL (2), compose stack and migrate-on-start instead of a migration job (3), local
go-common (4), `/api` instead of `/v1` for business routes and health (5), English-only catalog (6),
example removal (7).

Template documents that will contradict the project after this feature and must be updated in the
same change (not deviations, just stale text):

- `frontend/docs/code-conventions.md` §16 says "There is no test runner configured" (exception 2).
- `frontend/docs/adding-a-feature.md` and §12 of the conventions register pages under
  `PrivateRoute` → `AuthenticatedShell` (exception 1).
- `backend/.gitlab-ci.yml`, `backend/Dockerfile` reference `CI_JOB_TOKEN` / `GOPRIVATE` for the
  private module (exception 4). `backend/values.yml` (Helm) is unused locally and is left untouched;
  noted in `docs/KNOWN_LIMITATIONS.md`.

---

## R1. Go module name and local go-common

**Decision**: Rename the module `backend-template` → `smart-grid-asset-registry` (go.mod, every
import, `Makefile` `MODULE`, `.mockery.yaml`). Re-implement go-common under
`backend/pkg/gocommon/` keeping the upstream package paths below it
(`pkg/gocommon/pkg/apperror`, `pkg/gocommon/response`, `pkg/gocommon/pkg/logger`,
`pkg/gocommon/pkg/database`, `pkg/gocommon/pkg/utils`, `pkg/gocommon/pkg/middleware/fiber`).
Only the symbols the remaining code uses are implemented, with the v1.0.0 signatures:
`apperror.{AppError,New,BadRequest,NotFound,Forbidden,Conflict}`,
`response.{Response,Common,Success,Error}`, `logger.{New,WithCtx}`,
`database.{DBConfig,Connect}`, `utils.{NewErrorHandler,NotFoundHandler,RegisterHealthCheck,NewShutdownHelper}`,
`middleware/fiber.InstallCommonMiddleware`. The single addition (exception 4): `AppError` gains an
`ErrorCode string` field, `internal/apperr` fills it, and `NewErrorHandler` emits
`"data": {"errorCode": "<CODE>"}` when it is set (otherwise `"data": null`, as upstream).

**Rationale**: Same names and signatures let the template code and rules apply unchanged apart from
the import path. Implementing only what is used keeps dead code out (frontend B34 spirit, Go
`unused` lint).

**Note — provenance**: the private v1.0.0 source is present in the local Go module cache. It is
company code, so it is used only to confirm signatures and observable behaviour (JSON envelope,
status mapping); the implementation is written fresh. The user confirmed (2026-11-19) that this
use of the cached source is acceptable; copying it is not.

**Alternatives**: vendoring the cached module (licence/credentials problem, ships unused code);
replacing go-common calls with Fiber primitives (breaks "same names and signatures").

## R2. Where the import rules live (pure package)

**Decision**: `backend/domain/assetimport/` (exception 10) with:
`parse.go` (CSV → rows + file-level error), `normalize.go` (trim, blank → absent),
`field_rules.go` (FR-005…FR-010), `hierarchy_rules.go` (FR-011…FR-016a), `reimport_rules.go`
(FR-018/019), `serial_warnings.go` (FR-020), `validate.go` (orchestration, deterministic output),
`types.go`. Input: parsed rows + `StoredAssets` (map `asset_id → *entity.Asset`, plus a serial-number
index). Output: `Report` with accepted assets **in insert order**, unchanged rows, rejected rows
(sorted by row number), warnings, counts. Issues carry `Code` + `Args`, not text.

**Rationale**: The user requires the rules to be pure and fed a map of stored assets. `domain/` is
the only template layer that already forbids framework imports, and its tests need no mocks. Codes
(not messages) keep the engine independent of the English catalog.

**Alternatives**: `internal/usecase/asset_import/helpers.go` (one huge file, package imports logger
and apperr, so not pure); a new `internal/importrules` (internal packages may import frameworks;
weaker guarantee).

## R3. Evaluation algorithm (order independence)

**Decision**: Two passes over an index, never over the row sequence:
1. *Field pass* per row (independent): required fields, enums, numbers, dates, duplicate IDs
   (counted over the whole file first).
2. *Hierarchy pass* only for rows whose fields are valid: parent presence/self/not-found, parent
   type (in-file row or stored asset), cycles found by following parent pointers over the map of
   unique in-file IDs (every node on a cycle gets `CYCLE_DETECTED`), then `PARENT_REJECTED`
   propagated to a fixpoint (skipping cycle members), then re-import comparison against stored
   assets, then serial-number warnings over accepted + stored assets.
Accepted assets are emitted parents-first (depth, then asset ID) so a batched insert never violates
the parent foreign key. Message arguments that list IDs or rows are sorted.

**Rationale**: Every decision depends only on sets and maps, so permutations cannot change it
(SC-002). Running the rules on the sample with a throwaway script confirmed stability across 20
shuffles (see R12).

**Alternatives**: single pass with deferred parents (order-sensitive bugs are easy); DB-side
recursive CTE validation (rules leave the backend's pure layer; contradicts Principle II testing).

## R4. Validate-then-confirm flow and endpoints

**Decision**: Two endpoints, both `multipart/form-data` with parts `file` and `mode`
(`PARTIAL` default | `STRICT`):
- `POST /api/imports/validate` → `200`, outcome `VALIDATED_ONLY`; no transaction, no lock, no record.
- `POST /api/imports` → `201` with the created import record + report for every *completed*
  attempt (`COMMITTED`, `NOTHING_TO_COMMIT`, `STRICT_BLOCKED`), because each one stores an import
  record (FR-029). The browser resends the same `File` object it kept from the validate step;
  the server re-validates (FR-022).
File-level problems → `422` with a specific `data.errorCode`; bad `mode`/missing part → `400
INVALID_REQUEST`; unexpected failure → `500`, transaction rolled back, nothing stored, logged
(FR-030). Details in [contracts/imports-api.md](contracts/imports-api.md).

**Rationale**: Stateless server (no preview tokens or temp files), matches the spec's "re-validate
at confirm time", and keeps the template's "POST create → 201" rule. The static `/validate` route is
registered before any parameterised route (`03-handler.md`).

**Alternatives**: one endpoint with `?dryRun=true` (mixes 200/201 semantics on one route); server
-side preview token + temp storage (state, expiry, storage rules of D6).

## R5. Transaction and concurrency

**Decision**: `CreateAssetImport` runs everything in one `txManager.WithinTx`: first
`assetImportRepository.AcquireLock` (`SELECT pg_advisory_xact_lock(?)` with a fixed key), then the
stored-asset lookups, the pure validation, the import record insert, and the batched asset insert.
Strict-blocked and nothing-to-commit attempts still commit the transaction (they store a record).
The lock is released automatically at commit/rollback.

**Rationale**: Read-then-write must be atomic (`04-usecase.md`); the advisory lock serialises
imports (FR-024) without locking the `asset` table for readers. Unique and FK constraints remain
the final guarantee.

**Alternatives**: `SERIALIZABLE` isolation (retries needed); `LOCK TABLE asset` (blocks later
browse/search readers); optimistic insert-and-catch (unfriendly errors, partial reasoning).

## R6. Loading stored assets

**Decision**: Two repository reads, each with a single array parameter:
`FindByAssetIDs(ids)` for every asset ID and parent ID in the file, and
`FindBySerialNumbers(serials)` for every non-blank serial. SQL uses `= ANY(?)` with a `[]string`
so a 5 MB file (~50 000 rows) never hits PostgreSQL's 65 535 bind-parameter limit. Results become the
`StoredAssets` map handed to the pure package. Inserts use `CreateInBatches` of 1 000 rows.

**Rationale**: Stored assets already satisfy every hierarchy rule, so a stored parent is enough to
prove a chain reaches a substation; only stored rows the file mentions are needed. This keeps the
"map of stored assets" contract while scaling past a full-table read.

**Alternatives**: `FindAll` (simple, grows with the registry); `IN ?` expansion (parameter limit).

## R7. Schema choices

**Decision** (full detail in [data-model.md](data-model.md)): tables `asset_import` and `asset`;
UUID primary keys generated in the usecase; `asset.asset_id` unique; `asset.parent_asset_id`
references `asset(asset_id)`; `asset.import_id` references `asset_import(id)`; free text as `TEXT`;
numbers as unconstrained `NUMERIC`; dates as `DATE`; enums as `VARCHAR`; blanks stored as `NULL`;
no CHECK, no triggers (exception 8). Both tables keep the template audit columns
`created_at`, `created_by_name`, `updated_at`, `updated_by_name` (D2 rejected); with no users,
the `*_by_name` columns hold the fixed value `"system"`, defined once as
`usecase.SystemActorName` in `domain/usecase/asset_import.go` and set by the usecase like the
template's `ActorName`. The column names end in `ByName`, so the go-common audit plugin's
`CreatedBy`/`UpdatedBy` handling does not touch them (rule 05). `asset.import_id` stays as well:
it records which import created the asset, which the actor name cannot.

**Rationale**: The natural key is what the CSV and the later browse URLs use, so the self-reference
on `asset_id` keeps parent links readable and needs no ID translation. `TEXT` avoids an unspecified
"field too long" rule that would otherwise surface as a 500. Unconstrained `NUMERIC` keeps the value
exactly as typed, so "unchanged" comparisons do not suffer rounding.

**Alternatives**: `parent_id UUID → asset(id)` (extra lookups, less readable); `VARCHAR(255)` +
a new length rule (not in the spec); dropping the `*_by_name` columns (proposed as D2, rejected).

## R8. Migrations

**Decision**: Keep the template format (`NNNNNN_description.up.sql` / `.down.sql`, idempotent).
Delete the never-deployed `000001_create_product.*` in the foundation phase; add
`000001_create_asset_import` and `000002_create_asset`. `backend/migrations/migrations.go`
embeds `*.sql` (`//go:embed`), and `cmd/api/main.go` runs golang-migrate (`iofs` source,
`pgx/v5` database driver) before Fiber starts; failure is fatal (`log.Fatal` is allowed in `main`).
Integration tests apply the same embedded migrations to the test database.

**Rationale**: User decision; exception 3. Nothing was ever released, so removing the product
migration does not break the "never edit a released migration" rule, and it avoids creating a
useless `product` table on every start while the product Go code still exists as a reference.

**Alternatives**: separate migrate container (two moving parts); GORM AutoMigrate (no down
migrations, not the template format).

## R9. Error codes and messages

**Decision**: All stable codes (file-level, row-level, warning) live in `domain/errorcode` (+ `All`)
and have one English message each in the `internal/apperr` catalog. `apperr` gains an exported
`Message(code string, args ...any) string` that the handler uses to render row issues; business
errors keep using `BadRequest` / `UnprocessableEntity` etc., which now also set `ErrorCode`.
`TestCatalog_EveryErrorCodeHasAMessage` covers the new codes automatically.

**Rationale**: One catalog (exception 6), one test guarding completeness, and the pure package
stays text-free.

**Alternatives**: a second catalog inside `domain/assetimport` (two sources of truth).

## R10. Frontend shape

**Decision**:
- Remove auth (exception 1): `modules/auth`, `modules/users`, `pages/auth`, `pages/no-access`,
  `PrivateRoute`, token storage use, `UNAUTHENTICATED` / `FORBIDDEN` handling in `baseQuery`.
  `AuthenticatedShell` becomes `AppShell` (same shape, no account area).
- `toServerError` is adapted to the go-common envelope: `code` = `data.errorCode` when present,
  else derived from the HTTP status (`500 → INTERNAL_SERVER_ERROR`, `429`, `413 → PAYLOAD_TOO_LARGE`).
  The conventions explicitly allow this ("Adjust `ErrorCode` and `toServerError` if your API
  differs").
- Main page `pages/assets` at `AppRoute.ASSETS = "/assets"` (root redirects there). It holds an
  empty state for now; browsing arrives in a later feature. The import UI is a panel over it,
  open when the URL has `?panel=import` (Principle V).
- New generic `libs/components/panel` built on the native `<dialog>` element with `showModal()`:
  the browser makes the rest of the page inert (focus trap) and fires `cancel` on Escape; the panel
  turns that into removing the search parameter, and returns focus to the element that opened it.
- `modules/imports`: `imports-api.ts` with `validateImport` and `createImport` mutations sending
  `FormData`, DTO types written by hand from the contract, `ImportMode` / `ImportOutcome` enums,
  and a Zod schema that only checks a file is selected and the mode is one of two values (no size,
  type, or content rules — Principle II).
- Results come from the mutation state (no copy into local state). Choosing another file resets
  both mutations so a stale result is never shown next to a new file.

**Rationale**: Mirrors `modules/notes` / `pages/notes`; `<dialog>` gives correct focus handling with
no dependency.

**Alternatives**: a separate `/import` route (loses the "panel over the main page" decision);
a hand-written focus trap (more code, more a11y risk); a dialog library (new dependency).

## R11. Frontend tests

**Decision**: Add `vitest`, `jsdom`, `@testing-library/react`, `@testing-library/user-event`,
`@testing-library/jest-dom` as pinned dev dependencies (exception 2; the last three are the
standard parts of "React Testing Library"). `test` block in `vite.config.ts` (jsdom, setup file in
`src/libs/modules/testing/`), `"test": "vitest run"` script, tests named `*.test.ts(x)` next to the
code, `fetch` mocked with `vi.stubGlobal`. `knip` is told about the Vitest entry points.

**Rationale**: Vitest reuses the Vite config and aliases; no Jest/Babel duplication.

**Alternatives**: Jest (second transform pipeline); Playwright (heavier, not requested).

## R12. Golden test for the sample file

**Decision**: Copy the sample to `backend/domain/assetimport/testdata/grid_assets.csv`. The golden
expectations (`testdata/grid_assets.golden.json`: per-row codes, counts, warnings) are written when
the test is written, checked by hand, then frozen. A throwaway script applying the spec's rules
produced these **provisional** values for that hand check:

- 222 data rows → 205 accepted, 0 unchanged (empty store), 17 rejected; 13 accepted substations.
- Rejected: 91 PNL-M `PARENT_NOT_FOUND`; 92 SWB-S `PARENT_NOT_FOUND`; 93 SWB-CY-A, 94 SWB-CY-B,
  95 SWB-CY-C, 96 PNL-CY-C `CYCLE_DETECTED` + `INVALID_PARENT_TYPE`; 97 PNL-CY-D `CYCLE_DETECTED`;
  98 PNL-BP, 99 TX-BP `INVALID_PARENT_TYPE`; 178 GEN-001 `INVALID_ASSET_TYPE`; 179 TX-BD
  `INVALID_DATE`; 180 TX-NR `NEGATIVE_RATING`; 181 PNL-BS `INVALID_OPERATIONAL_STATUS`;
  182 (no ID) `ASSET_ID_REQUIRED`; 183 PNL-N `ASSET_NAME_REQUIRED`; 184 SS-H-P
  `SUBSTATION_HAS_PARENT`; 185 TX-N-P `PARENT_REQUIRED`.
- Warnings: 89 TX-D1 and 90 TX-D2 `DUPLICATE_SERIAL_NUMBER`.
- The sample triggers no `PARENT_REJECTED`, `DUPLICATE_ASSET_ID` or `INVALID_NUMBER`; table tests
  cover those.

The brief's "220 valid assets" does not match 205; "13 substations" does. Recorded in
`docs/KNOWN_LIMITATIONS.md`.

## R13. Local stack and tests in containers

**Decision**: Root `docker-compose.yml`:
- `db`: `postgres:16-alpine`, host port `55432`, named volume, health check, an init script that
  also creates the `asset_registry_test` database.
- `api`: built from `backend/Dockerfile` (private-module lines removed), `APP_PORT=8080`, depends on
  healthy `db`, runs migrations on start.
- `web`: built from `frontend/Dockerfile` with `VITE_APP_API_ORIGIN_URL=/api`; nginx on `3000`
  forwards `/api` to `api:8080`.
- Profile `test`: `backend-test` (Go image, `make check` with
  `TEST_DATABASE_URL=postgres://…@db:5432/asset_registry_test`) and `frontend-test` (Node image,
  `pnpm install --frozen-lockfile && pnpm test`).
Root `.gitattributes`: `* text=auto eol=lf`, binaries (`*.woff2`, `*.ico`) marked `binary`.
Fiber `BodyLimit` raised to 6 MB so a 5 MB file plus multipart overhead reaches the usecase, which
enforces the 5 MB limit with a proper error code.

**Rationale**: Exception 3 and user decisions; integration tests use a separate database so they
can truncate tables without wiping the developer's data.

**Alternatives**: testcontainers (needs Docker-in-Docker inside the test profile); one shared
database (tests would destroy local data).

## R14. Logging

**Decision**: Usecases follow `09-logging-observability.md`: `log := logger.WithCtx(...).WithField("usecase", ...)`;
one `Info("success")` with `fileName`, `mode`, `outcome`, counts and `importID`; `Error` with
`WithError` for technical failures inside the transaction (FR-030 "failed attempts are logged");
file-level `422`s are not logged as errors. Row contents are never logged.

**Rationale**: Template rule; brief weights logging.

**Alternatives**: per-row logs (noise, data in logs).
