---
description: "Task list for the csv-import feature"
---

# Tasks: CSV Import of Grid Assets

**Input**: Design documents from `specs/001-csv-import/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md),
[data-model.md](data-model.md), [contracts/imports-api.md](contracts/imports-api.md),
[contracts/ui-import-panel.md](contracts/ui-import-panel.md), [quickstart.md](quickstart.md),
constitution 1.1.0

**Tests**: Required. Constitution IV: every import and hierarchy rule gets a table-driven test
**written and failing before** the rule is implemented (separate test task first); the sample CSV
gets a golden test; repositories and usecases are also tested against real PostgreSQL. Template
granularity: one task per use case, handler, repository method and migration, each delivered with
its test file.

**Organization**: Setup → Foundational → one phase per user story (spec priority order) → Polish.
Every user-story phase ends with a checkpoint run in the real stack (`docker compose up --build`).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: can run in parallel (different files, no dependency on an unfinished task)
- **[Story]**: US1–US4 from spec.md
- Paths are relative to the repository root. Backend Go module path after T002:
  `smart-grid-asset-registry`.

## Reference shapes (copy these, do not invent)

| New                        | Copy the shape of                                                                                   |
| -------------------------- | --------------------------------------------------------------------------------------------------- |
| value objects              | `backend/domain/valueobject/product_status.go` (+ `_test.go`)                                       |
| usecase package            | `backend/internal/usecase/product/{init,helpers,create_product}.go` (+ tests, `mocks.StubWithinTx`) |
| handler package            | `backend/internal/handler/product/{init,helpers,create_product,testutil_test}.go`                   |
| repository package         | `backend/internal/repository/psql/product/{init,create,find_by_code,testdb_test}.go`                |
| model / mapper             | `backend/internal/repository/psql/{model,mapper}/product.go`                                        |
| frontend feature module    | `frontend/src/modules/notes/`                                                                       |
| frontend page + components | `frontend/src/pages/notes/` (list states in `notes-list.tsx`, form in `note-create-form.tsx`)       |

The `product` and `notes` examples stay in the tree as references until T108/T109 delete them.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: The project foundation required by constitution exceptions 1–7 and the user's
setup decisions: module rename, local go-common, auth removal, `/api`, compose, `.gitattributes`,
Vitest.

- [x] T001 [P] Create `.gitattributes` at the repository root: `* text=auto eol=lf`, and mark `*.woff2`, `*.ico` and `*.png` as `binary` (research R13); renormalize with `git add --renormalize .`
- [x] T002 Rename the Go module `backend-template` → `smart-grid-asset-registry` in `backend/go.mod`, every import in `backend/**/*.go`, `MODULE :=` in `backend/Makefile` and both package paths in `backend/.mockery.yaml`; `go build ./...` must pass (research R1)
- [x] T003 [P] Local go-common `apperror` in `backend/pkg/gocommon/pkg/apperror/error.go` + `error_test.go`: `AppError{Code int; Message string; Errors interface{}; ErrorCode string}` (the `ErrorCode` field is the only addition, exception 4), `Error()`, `New`, `NotFound`, `BadRequest`, `Forbidden`, `Conflict` with the v1.0.0 signatures (cached source used for signatures only, never copied — research R1)
- [x] T004 [P] Local go-common `response` in `backend/pkg/gocommon/response/response.go` + `response_test.go`: generic `Response[T]{Code, Status, Message, Data (no omitempty), Meta (omitempty)}`, `Common = Response[any]`, `Success(c, code, message, data)` (`status: "success"`) and `Error(c, code, message, data)` (`status: "error"`); test that `"data": null` and `"data": []` are always emitted
- [x] T005 [P] Local go-common `logger` in `backend/pkg/gocommon/pkg/logger/logger.go` + `logger_test.go`: `New(isProd bool) *logrus.Logger` (JSON in prod, text otherwise) and `WithCtx(ctx, log) *logrus.Entry` adding the request ID from the context when present
- [x] T006 [P] Local go-common `database` in `backend/pkg/gocommon/pkg/database/{config.go,database.go}`: `DBConfig{Host, User, Password, DBName, Port, MaxIdleConns, MaxOpenConns, ConnMaxLifetime}` and `Connect(cfg *DBConfig, log *logrus.Logger, isProd bool) (*gorm.DB, error)` with pool settings; no audit plugin (no `CreatedBy`/`UpdatedBy` fields exist; the `*ByName` columns are set by usecases). Update the comment in `backend/cmd/api/main.go` that mentions the audit plugin
- [x] T007 Local go-common `utils` in `backend/pkg/gocommon/pkg/utils/{error.go,health.go,shutdown.go}` + `error_test.go`, `health_test.go`: `NewErrorHandler(log) fiber.ErrorHandler` (`*AppError` → its status and message, `data` = `{"errorCode": "<CODE>"}` when `ErrorCode` is set, else `null`; `*fiber.Error` → its status; anything else → `500 "Internal Server Error"`; log ≥ 500), `NotFoundHandler`, `RegisterHealthCheck(router, db, rdb)` (pings the DB, `rdb` may be nil), `NewShutdownHelper(log)` with `Wait()` and `Graceful(map[string]func(ctx) error)` (depends on T003–T005)
- [x] T008 [P] Local go-common `middleware/fiber` in `backend/pkg/gocommon/pkg/middleware/fiber/{common.go,identity.go}` + `common_test.go`: `InstallCommonMiddleware(app, log)` = request ID, access log, security headers. `identity.go` holds only `UserIdentity`, `UserIdentityKey` and `GetUserIdentity`, still used by the `product` example handler; T108 deletes `identity.go`
- [x] T009 Point every go-common import in `backend/` at `smart-grid-asset-registry/pkg/gocommon/...`, drop `repo.smf-indonesia.co.id/...` from `backend/go.mod` / `go.sum` (`go mod tidy`), and remove the `CI_JOB_TOKEN` / `GOPRIVATE` / private-module lines from `backend/Dockerfile` and `backend/.gitlab-ci.yml` (exception 4); `go build ./... && go vet ./...` pass without company credentials (depends on T002–T008)
- [x] T010 Update `backend/internal/apperr/apperr.go` + `apperr_test.go`: catalog entries hold English text only (drop `ID`, exception 6); every constructor (`BadRequest`, `Forbidden`, `NotFound`, `Conflict`, `UnprocessableEntity`) sets `AppError.ErrorCode = code`; add exported `Message(code string, args ...any) string` used to render row issues (research R9); tests cover `ErrorCode` and `Message`
- [x] T011 [P] Remove `INTERNAL_SERVICE_SECRET` and `JWT_SECRET` from `backend/internal/config/config.go` (struct, `Load`, `Validate`), default `APP_PORT` to `8080`, update `backend/internal/config/config_test.go` and `backend/.env.example` (exception 1, 3)
- [x] T012 Rewrite `backend/internal/routes/router.go` to a public `/api` group (exception 1, 5): `utils.RegisterHealthCheck(api, deps.DB, nil)` at `/api/health`, no auth middleware chain, product routes no longer registered; remove the product wiring from `backend/internal/bootstrap/setup.go`; set `BodyLimit: 6 * 1024 * 1024` in `backend/internal/bootstrap/fiber.go` (research R13)
- [x] T013 Delete the never-deployed `backend/migrations/000001_create_product.up.sql` and `.down.sql` (research R8). The product Go code and its sqlmock tests stay until T108
- [x] T014 Remove frontend auth (exception 1): delete `frontend/src/modules/auth/`, `frontend/src/modules/users/`, `frontend/src/pages/auth/`, `frontend/src/pages/no-access/`, `frontend/src/pages/dashboard/` (it reads the signed-in user), `frontend/src/libs/components/private-route/`, `frontend/src/libs/helpers/check-is-auth-path/`, `frontend/src/libs/constants/bearer.constant.ts`, `frontend/src/assets/img/no-access.svg`; update `frontend/src/index.tsx` (no `PrivateRoute`, no auth routes; the index route **temporarily redirects to `AppRoute.NOTES`** until T042 switches it to `/assets`; keep `NOTES` registered until T109), `frontend/src/libs/components/app/app.tsx` (no authenticated-user query), `frontend/src/libs/enums/app-route.enum.ts`, barrels `helpers.ts` / `constants.ts`, and `StorageKey` (drop `TOKEN`)
- [x] T015 Rename `frontend/src/libs/components/authenticated-shell/` → `frontend/src/libs/components/app-shell/` (`AppShell`, same layout, no account area or sign-out button, `ShellLabel` without sign-out); update `frontend/src/index.tsx` and the navigation items constant (depends on T014)
- [x] T016 Adapt the API error layer to the go-common envelope (research R10): `frontend/src/libs/enums/error-code.enum.ts` (drop `UNAUTHENTICATED`, `FORBIDDEN`, `VALIDATION_FAILED`; add `PAYLOAD_TOO_LARGE`), `frontend/src/libs/modules/api/libs/helpers/to-server-error.helper.ts` (code = `data.errorCode` when present, else from HTTP status: `500` → `INTERNAL_SERVER_ERROR`, `429` → `TOO_MANY_REQUESTS`, `413` → `PAYLOAD_TOO_LARGE`, `404` → `NOT_FOUND`; message from `message`), `base-query.helper.ts` (no token header, no auth/forbidden branches), and the server error types in `frontend/src/libs/types/` and `frontend/src/libs/modules/api/libs/types/`; remove now-unused validation-error helpers and `use-server-form-errors` only if knip reports them unused (depends on T014)
- [x] T017 [P] Point the frontend at `/api`: `frontend/.env.example` (`VITE_APP_API_ORIGIN_URL=/api`, `VITE_APP_PROXY_SERVER_URL=http://localhost:8080`) and `frontend/nginx/nginx.conf` (`upstream backend { server api:8080; }`) (exception 3, 5)
- [x] T018 Add Vitest (exception 2, research R11): pinned dev deps `vitest`, `jsdom`, `@testing-library/react`, `@testing-library/user-event`, `@testing-library/jest-dom` in `frontend/package.json` + `pnpm-lock.yaml`; `"test": "vitest run"`; `test` block in `frontend/vite.config.ts` (environment `jsdom`, setup file); `frontend/src/libs/modules/testing/setup-tests.ts` (jest-dom matchers, `fetch` reset, and a `HTMLDialogElement.prototype.showModal`/`close` shim, since jsdom has no `<dialog>` modal support); register the Vitest config and test files as entries in `frontend/knip.config.ts`
- [x] T019 [P] Tests for the error layer in `frontend/src/libs/modules/api/libs/helpers/to-server-error.helper.test.ts`: envelope with `data.errorCode`, envelope with `data: null` and status 500/413/429, network failure (depends on T016, T018)
- [x] T020 [P] Root `docker-compose.yml` + `docker/postgres/init/01-create-test-db.sql` (research R13, exception 3): `db` = `postgres:16-alpine`, port `55432:5432`, named volume, `pg_isready` health check, database `asset_registry`, init script creates `asset_registry_test`; `api` = build `backend/`, `APP_PORT=8080`, DB env, `depends_on: db: condition: service_healthy`, port `8080:8080`; `web` = build `frontend/` with `VITE_APP_API_ORIGIN_URL=/api`, port `3000:80`; profile `test` with `backend-test` (Go 1.26 image, `make check` with `TEST_DATABASE_URL=postgres://postgres:postgres@db:5432/asset_registry_test?sslmode=disable`) and `frontend-test` (Node 24 image, `corepack enable && pnpm install --frozen-lockfile && pnpm test`)
- [x] T021 Setup checkpoint: `docker compose up --build` starts all three services; `curl -s localhost:8080/api/health` → 200 with the go-common envelope; `http://localhost:3000` redirects to `/notes` inside `AppShell` without any sign-in; `make check` and `make lint` (in `backend/`) and `pnpm lint`, `pnpm build`, `pnpm test` (in `frontend/`) exit 0. Record the commands and exit codes

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Domain types, error codes, schema, migrations on start, repository/usecase
contracts, test helpers and the frontend panel shell that every story builds on.

**⚠️ CRITICAL**: No user-story work starts before this phase is complete.

- [x] T022 [P] `backend/domain/valueobject/asset_type.go` + `asset_type_test.go`: `AssetType` with values `SUBSTATION`, `TRANSFORMER`, `LV_BOARD`, `SWITCHBOARD`, `SWITCHBOARD_PANEL`; `Parse`/`Scan`/`Value`/`String`/`IsValid`; `PermittedParent() (AssetType, bool)` — "`TRANSFORMER`, `LV_BOARD`, `SWITCHBOARD` → `SUBSTATION`; `SWITCHBOARD_PANEL` → `SWITCHBOARD`; `SUBSTATION` → none" (data-model.md)
- [x] T023 [P] `backend/domain/valueobject/operational_status.go` + `operational_status_test.go`: `IN_SERVICE`, `MAINTENANCE`, `OUT_OF_SERVICE` (case-sensitive; entity field is a pointer, `NULL` = "no status")
- [x] T024 [P] `backend/domain/valueobject/import_mode.go` + `import_mode_test.go`: `PARTIAL`, `STRICT`; "default `PARTIAL`"
- [x] T025 [P] `backend/domain/valueobject/import_outcome.go` + `import_outcome_test.go`: `VALIDATED_ONLY`, `COMMITTED`, `NOTHING_TO_COMMIT`, `STRICT_BLOCKED` ("`VALIDATED_ONLY` is never stored")
- [x] T026 Add every stable code to `backend/domain/errorcode/errorcode.go` (constants + `All`) and one English message each to the catalog in `backend/internal/apperr/apperr.go`: file-level `IMPORT_FILE_TOO_LARGE`, `IMPORT_FILE_EMPTY`, `IMPORT_FILE_NOT_CSV` (message names the line), `IMPORT_MISSING_COLUMNS`, `IMPORT_DUPLICATE_COLUMNS`; row-level `MALFORMED_ROW`, `ASSET_ID_REQUIRED`, `ASSET_TYPE_REQUIRED`, `ASSET_NAME_REQUIRED`, `DUPLICATE_ASSET_ID`, `INVALID_ASSET_TYPE`, `INVALID_OPERATIONAL_STATUS`, `INVALID_NUMBER`, `NEGATIVE_RATING`, `INVALID_DATE`, `SUBSTATION_HAS_PARENT`, `PARENT_REQUIRED`, `SELF_PARENT`, `PARENT_NOT_FOUND`, `INVALID_PARENT_TYPE`, `CYCLE_DETECTED`, `PARENT_REJECTED`, `ASSET_ID_CONFLICT`; warning `DUPLICATE_SERIAL_NUMBER`. Messages are readable, name the offending value through `%s` args and never contain DB details; `TestCatalog_EveryErrorCodeHasAMessage` passes (depends on T010)
- [x] T027 Map Fiber's 413 to a stable code (spec FR-003, research R4, analysis N1): in the local go-common error handler `backend/pkg/gocommon/pkg/utils/error.go`, a `*fiber.Error` with status 413 is answered with 413, `PayloadTooLargeMessage` and `data.errorCode` `PayloadTooLargeErrorCode` (`IMPORT_FILE_TOO_LARGE`), so `internal/apperr` does not depend on Fiber; case in `error_test.go`; `backend/internal/apperr/apperr_test.go` asserts the code and message equal `errorcode.ImportFileTooLarge` and its catalog message (depends on T012, T026)
- [x] T028 [P] `backend/domain/entity/asset.go`: `Asset{ID uuid.UUID; AssetID string; ParentAssetID *string; AssetType valueobject.AssetType; AssetName string; OperationalStatus *valueobject.OperationalStatus; VoltageKV, RatingKVA *float64; Manufacturer, Model, SerialNumber *string; CommissionedDate *time.Time /* date only, UTC midnight */; ImportID uuid.UUID; CreatedAt time.Time; CreatedByName string; UpdatedAt time.Time; UpdatedByName string}` (`DifferingFields` comes in T098)
- [x] T029 [P] `backend/domain/entity/asset_import.go`: `AssetImport{ID; FileName; Mode valueobject.ImportMode; Outcome valueobject.ImportOutcome; TotalRows, ImportedRows, UnchangedRows, RejectedRows, WarningRows int; CreatedAt; CreatedByName; UpdatedAt; UpdatedByName}`
- [x] T030 [P] Migration `backend/migrations/000001_create_asset_import.up.sql` / `.down.sql` (idempotent, `IF NOT EXISTS`), exactly per data-model.md: `id UUID PRIMARY KEY`, `file_name TEXT NOT NULL`, `mode VARCHAR(16) NOT NULL`, `outcome VARCHAR(32) NOT NULL`, `total_rows`, `imported_rows`, `unchanged_rows`, `rejected_rows`, `warning_rows INTEGER NOT NULL`, `created_at TIMESTAMPTZ NOT NULL`, `created_by_name VARCHAR(255) NOT NULL`, `updated_at TIMESTAMPTZ NOT NULL`, `updated_by_name VARCHAR(255) NOT NULL`; index `ix_asset_import_created_at (created_at DESC)`; **no CHECK constraints, no triggers, no defaults for the audit names** (exception 8, data-model.md)
- [x] T031 Migration `backend/migrations/000002_create_asset.up.sql` / `.down.sql` per data-model.md: `id UUID PRIMARY KEY`, `asset_id TEXT NOT NULL`, `parent_asset_id TEXT NULL REFERENCES asset(asset_id) ON DELETE RESTRICT`, `asset_type VARCHAR(32) NOT NULL`, `asset_name TEXT NOT NULL`, `operational_status VARCHAR(32) NULL`, `voltage_kv NUMERIC NULL`, `rating_kva NUMERIC NULL`, `manufacturer TEXT NULL`, `model TEXT NULL`, `serial_number TEXT NULL`, `commissioned_date DATE NULL`, `import_id UUID NOT NULL REFERENCES asset_import(id)`, `created_at TIMESTAMPTZ NOT NULL`, `created_by_name VARCHAR(255) NOT NULL`, `updated_at TIMESTAMPTZ NOT NULL`, `updated_by_name VARCHAR(255) NOT NULL`; indexes `ux_asset_asset_id` (unique), `ix_asset_parent_asset_id`, `ix_asset_serial_number`, `ix_asset_import_id`; no CHECK, no triggers (depends on T030)
- [x] T032 Run migrations on start (exception 3, research R8): `backend/migrations/migrations.go` (`package migrations`, `//go:embed *.sql`, exported `FS`); add `github.com/golang-migrate/migrate/v4` (`source/iofs`, `database/pgx/v5`) to `backend/go.mod`; in `backend/cmd/api/main.go` apply `Up()` before Fiber starts, treat `migrate.ErrNoChange` as success, `log.Fatalf` on any other error (depends on T030, T031)
- [x] T033 [P] GORM models `backend/internal/repository/psql/model/asset.go` and `asset_import.go` (gorm column tags, `TableName()` `asset` / `asset_import`, `*ByName` audit fields as in `model/product.go`) (depends on T028, T029)
- [x] T034 [P] Mappers `backend/internal/repository/psql/mapper/asset.go`, `asset_import.go` + `asset_test.go`, `asset_import_test.go`: `ToModelAsset`, `ToEntityAsset`, `ToEntityAssets` (non-nil slice), `ToModelAssetImport`, `ToEntityAssetImport`; every field and nil pointers covered (depends on T033)
- [x] T035 [P] Repository interfaces: `backend/domain/repository/asset.go` — `FindByAssetIDs(ctx, trx Tx, assetIDs []string) ([]*entity.Asset, error)`, `FindBySerialNumbers(ctx, trx Tx, serialNumbers []string) ([]*entity.Asset, error)`, `CreateMany(ctx, trx Tx, assets []*entity.Asset) error`; `backend/domain/repository/asset_import.go` — `Create(ctx, trx Tx, assetImport *entity.AssetImport) error`, `AcquireLock(ctx, trx Tx) error` (depends on T028, T029)
- [x] T036 [P] `backend/domain/assetimport/types.go` (exception 10: standard library and `domain/*` imports only — **no `github.com/google/uuid`**): `Column` constants for the eleven CSV columns in CSV order, `Row{Number int; Values map[Column]string}`, `ParsedFile{Rows []Row}`, `FileError{Code string; Args []any}`, `StoredAssets{ByAssetID map[string]*entity.Asset; BySerial map[string][]string}`, `Issue{Code string; Args []any}`, `RejectedRow{Row int; AssetID *string; Errors []Issue}`, `Warning{Row int; AssetID string; Issue Issue}`, `Counts{Total, ToImport, Unchanged, Rejected, WarningRows int}`, `Report{Accepted []*entity.Asset /* ID, ImportID, audit fields left zero */; UnchangedRows []int; Rejected []RejectedRow; Warnings []Warning; Counts Counts}` (data-model.md) (depends on T028)
- [x] T037 Usecase contract `backend/domain/usecase/asset_import.go` + `asset_import_test.go`: `AssetImportUsecase{ValidateAssetImport(ctx, ValidateAssetImportInput) (ValidateAssetImportOutput, error); CreateAssetImport(ctx, CreateAssetImportInput) (CreateAssetImportOutput, error)}`; inputs `{FileName string; File io.Reader; Mode valueobject.ImportMode}` (no size field: `Parse` alone enforces the limit, data-model.md) with `Validate()` (file present, file name present, mode valid → plain errors for `400 INVALID_REQUEST`); outputs carry `Report assetimport.Report` and, for create, `Import *entity.AssetImport`; constants `SystemActorName = "system"` and `MaxFileBytes = 5 * 1024 * 1024`; table test for `Validate()` (depends on T036)
- [x] T038 Regenerate mocks with `make mocks` into `backend/internal/mocks/repository/` and `backend/internal/mocks/usecase/` (new `AssetRepository`, `AssetImportRepository`, `AssetImportUsecase`) (depends on T035, T037)
- [x] T039 [P] `backend/internal/testsupport/pgtest/pgtest.go` (exception 11): `Open(t) *gorm.DB` skips the test when `TEST_DATABASE_URL` is unset, applies the embedded migrations (`migrations.FS`), and registers a cleanup that truncates `asset` and `asset_import`; `pgtest_test.go` checks the skip path (depends on T032)
- [x] T040 [P] Generic panel `frontend/src/libs/components/panel/panel.tsx` + `styles.module.css` + `panel.test.tsx` (Principle V, research R10): native `<dialog>` opened with `showModal()` when `isOpen`, `aria-labelledby` its title, **Close** button, `cancel` event (Escape) → `onClose`, focus returned to the element focused before opening; tests: opens/closes with `isOpen`, Escape calls `onClose`, Close button calls `onClose`, focus restored. The real focus trap comes from the browser's modal `<dialog>` and cannot be exercised under the jsdom shim; it is verified by hand in checkpoint T076 (quickstart UI step 2) (depends on T018)
- [x] T041 [P] Panel URL state: enums `frontend/src/libs/enums/search-parameter.enum.ts` (`PANEL: "panel"`) and `panel-name.enum.ts` (`IMPORT: "import"`) in the `enums.ts` barrel; hook `frontend/src/libs/hooks/use-panel/use-panel.hook.ts` + `use-panel.hook.test.tsx` returning `{ openPanel, closePanel, isOpen }` for a given `PanelName` through `useSearchParams` (open pushes history, so Back closes; unknown values = closed) (depends on T018)
- [x] T042 Main page `frontend/src/pages/assets/assets.tsx` + `styles.module.css` + `libs/enums/assets-page-message.enum.ts` + `assets.test.tsx`: `PageIntro`, empty state ("No assets yet. Import a CSV file to get started."), **Import CSV** button calling `openPanel(PanelName.IMPORT)`, `Panel` titled "Import assets from CSV" with a placeholder body; add `AppRoute.ASSETS = "/assets"`, switch the `AppRoute.ROOT` index redirect from `/notes` (T014) to `/assets` in `frontend/src/index.tsx`, navigation item in `AppShell`; tests: Import CSV sets `?panel=import`, deep link opens the panel (depends on T015, T040, T041)
- [x] T043 Feature module skeleton `frontend/src/modules/imports/` per [contracts/imports-api.md](contracts/imports-api.md): `APIPath.IMPORTS = "/imports"`; `libs/enums/` `imports-api-path.enum.ts` (`ROOT: "/"`, `VALIDATE: "/validate"`), `import-mode.enum.ts`, `import-outcome.enum.ts`, `import-validation-message.enum.ts`; `libs/types/` one file each for `ImportResultDto`, `ImportCountsDto`, `ImportRejectedRowDto`, `ImportIssueDto`, `ImportWarningDto`, `ImportRecordDto`, `ImportRequestDto` (`z.infer`); `libs/validation-schemas/import-file.validation-schema.ts` — `file` is a `File`, `mode` is `PARTIAL | STRICT`, **nothing else** (no size, type or content rules — Principle II); barrel `imports.ts` (endpoints are added per story)

**Checkpoint**: `make check`, `make lint`, `pnpm lint`, `pnpm build`, `pnpm test` exit 0;
`docker compose up --build` applies migrations 000001–000002 on start (`\dt` in `asset_registry`
shows `asset`, `asset_import`, `schema_migrations`); `/` redirects to `/assets`, the empty state
shows, and the panel opens and closes via URL, button and Escape.

---

## Phase 3: User Story 1 — Validate a file and see every problem (Priority: P1) 🎯 MVP

**Goal**: An operator checks a CSV and sees total / to-import / unchanged / rejected counts, every
rejected row with row number, asset ID, code and message, and warnings — with nothing stored.

**Independent Test**: Validate the sample on an empty store → 222 rows, 205 to import, 17
rejected with the codes in research R12, warnings on rows 89/90, "Validated only — nothing saved";
`asset` table still empty. Shuffled rows give the same IDs and codes.

### Tests for User Story 1 — write first, confirm they FAIL ⚠️

- [x] T044 [P] [US1] `backend/domain/assetimport/parse_test.go` (table-driven, FR-002–FR-004, edge cases): header resolved by name in any order; extra columns ignored; a header repeating a column name (e.g. two `asset_id`) → `FileError{IMPORT_DUPLICATE_COLUMNS, [repeated names]}`; UTF-8 BOM accepted; double-quoted fields with commas/newlines; missing columns → `FileError{IMPORT_MISSING_COLUMNS, [names in CSV order]}`; empty file and header-only → `IMPORT_FILE_EMPTY`; invalid UTF-8 → `IMPORT_FILE_NOT_CSV`; **any CSV syntax error** (unclosed quote, stray quote in an unquoted field) on any line → `IMPORT_FILE_NOT_CSV` with that line number as arg, and no rows returned; input longer than `maxBytes` → `IMPORT_FILE_TOO_LARGE`; a record with valid syntax but the wrong field count is kept as a row that later gets `MALFORMED_ROW` while other rows parse; row numbers are file line numbers with the header as 1
- [x] T045 [P] [US1] `backend/domain/assetimport/normalize_test.go`: leading/trailing whitespace trimmed in every field; whitespace-only = empty; inner whitespace kept ("Dual Node B")
- [x] T046 [P] [US1] `backend/domain/assetimport/field_rules_test.go` (one table case per code, FR-005–FR-010): `ASSET_ID_REQUIRED`, `ASSET_TYPE_REQUIRED`, `ASSET_NAME_REQUIRED`; `DUPLICATE_ASSET_ID` on **every** row sharing the ID; `INVALID_ASSET_TYPE` (`GENERATOR`, lower-case `substation`); blank status accepted, `BROKEN` → `INVALID_OPERATIONAL_STATUS`; `INVALID_NUMBER` for `voltage_kv` and `rating_kva` (arg = column); `NEGATIVE_RATING` for `-500`; negative `voltage_kv` and future dates accepted (clarification Q2); dates `15/2/09` → 2009-02-15, `2009-02-15` ok, `31/2/20` and `2026-13-41` → `INVALID_DATE`; a row with several field problems lists all of them; `MALFORMED_ROW`
- [x] T047 [P] [US1] `backend/domain/assetimport/hierarchy_rules_test.go` (FR-011–FR-016a): `SUBSTATION_HAS_PARENT`; `PARENT_REQUIRED`; `SELF_PARENT`; `PARENT_NOT_FOUND` (incl. `SWB-S` → `SWB-SELF`, which is not `SELF_PARENT`); parent stored in `StoredAssets` accepted; `INVALID_PARENT_TYPE` for an in-file parent and for a stored parent; parent with an unsupported type → `PARENT_REJECTED` only; `CYCLE_DETECTED` on every member of a 2-cycle and a 3-cycle; rows below a cycle → `PARENT_REJECTED`, cycle members never get `PARENT_REJECTED`; `PARENT_REJECTED` propagates down a 3-level chain; children of a `DUPLICATE_ASSET_ID` parent → `PARENT_REJECTED`; SWB-CY-A gets both `INVALID_PARENT_TYPE` and `CYCLE_DETECTED`; hierarchy rules skipped for rows with field errors; child row before its parent and grandparent is accepted
- [x] T048 [P] [US1] `backend/domain/assetimport/serial_warnings_test.go` (FR-020): two accepted rows sharing a serial both warn, naming each other; accepted row vs stored asset with another ID warns; rejected rows neither warn nor trigger warnings; blank serials ignored; args sorted; warnings never change acceptance
- [x] T049 [P] [US1] `backend/domain/assetimport/validate_test.go`: `Counts` invariant `Total = ToImport + Unchanged + Rejected`; `Rejected` sorted by row; `Accepted` ordered parents first (depth, then asset ID) so a batched insert satisfies the FK; every accepted asset has zero `ID`, `ImportID` and audit fields (the usecase assigns them); `TestValidate_OrderIndependent` — 20 seeded permutations of the sample give identical accepted IDs, rejected IDs and codes (SC-002, FR-017)
- [x] T050 [US1] Golden test (Principle IV, research R12): copy `D:\my-data\personal\Application\cmc\take-home_assignment_Oscar\grid_assets.csv` to `backend/domain/assetimport/testdata/grid_assets.csv` unchanged; write `backend/domain/assetimport/golden_test.go` comparing per-row codes, counts and warnings with `testdata/grid_assets.golden.json`; fill the JSON by hand from research R12 (222 / 205 / 0 / 17, rows 91–99 and 178–185, warnings 89–90) and freeze it only after checking each row by hand against the spec

### Implementation for User Story 1

- [x] T051 [US1] `backend/domain/assetimport/parse.go`: `Parse(r io.Reader, maxBytes int64) (ParsedFile, *FileError)` — the only place the 5 MB limit is enforced (reads at most `maxBytes + 1` bytes); `encoding/csv` with `FieldsPerRecord = -1`, `LazyQuotes = false`; a `*csv.ParseError` other than a field-count error → `IMPORT_FILE_NOT_CSV` with `ParseError.Line`; BOM strip, UTF-8 check, header resolution incl. duplicate detection — T044 passes
- [x] T052 [US1] `backend/domain/assetimport/normalize.go` — T045 passes
- [x] T053 [US1] `backend/domain/assetimport/field_rules.go` — T046 passes
- [x] T054 [US1] `backend/domain/assetimport/hierarchy_rules.go` (index-based two-pass algorithm, research R3) — T047 passes
- [x] T055 [US1] `backend/domain/assetimport/serial_warnings.go` — T048 passes
- [x] T056 [US1] `backend/domain/assetimport/validate.go`: `Validate(file ParsedFile, stored StoredAssets) Report` orchestrating field → hierarchy → serial warnings, building `Accepted` entities (blank → `nil`, dates parsed; `ID`, `ImportID` and audit fields left zero) — T049 and T050 pass; no file in `backend/domain/assetimport/` imports `github.com/google/uuid`
- [x] T057 [P] [US1] `backend/internal/repository/psql/asset/init.go` + `find_by_asset_ids.go` + `find_by_asset_ids_test.go` + `testdb_test.go` (package `asset_repository`): `WHERE asset_id = ANY(?)` with one `[]string` argument (research R6); sqlmock cases: success, empty input → no query + empty slice, DB error, `wrapTx`, `unsupportedTx`
- [x] T058 [US1] `backend/internal/repository/psql/asset/find_by_serial_numbers.go` + `find_by_serial_numbers_test.go`: `WHERE serial_number = ANY(?)`; same sqlmock cases (depends on T057 for `init.go`)
- [x] T059 [US1] Repository integration test `backend/internal/integrationtest/asset_repository_read_test.go` (exception 11): seed rows directly, `FindByAssetIDs` / `FindBySerialNumbers` return the right assets incl. nullable fields and dates; skipped without `TEST_DATABASE_URL` (depends on T039, T057, T058)
- [x] T060 [US1] **Test first** — `backend/internal/usecase/asset_import/helpers_test.go` (package `asset_import_usecase`), seen failing before T061: `parseInput` maps input `Validate()` failures to `apperr.BadRequest(INVALID_REQUEST)` and each `Parse` `FileError` (incl. `IMPORT_FILE_TOO_LARGE` from `MaxFileBytes`, `IMPORT_FILE_NOT_CSV` with line, `IMPORT_DUPLICATE_COLUMNS`) to `apperr.UnprocessableEntity(code, args...)`; `loadStoredAssets` asks the repository for asset IDs ∪ parent IDs and serials exactly once each and builds `StoredAssets`; repository errors are returned
- [x] T061 [US1] Usecase package `backend/internal/usecase/asset_import/init.go` + `helpers.go`: constructor `NewAssetImportUsecase(txManager, assetRepository, assetImportRepository, log)` with injectable `now` and `newID` (`uuid.New`); helpers `parseInput` (calls `assetimport.Parse(input.File, usecase.MaxFileBytes)`; no separate size check) and `loadStoredAssets(ctx, trx, file)` — T060 passes
- [x] T062 [US1] `backend/internal/usecase/asset_import/validate_asset_import.go` + `validate_asset_import_test.go`: log setup → parse → load stored (no transaction, no lock) → `assetimport.Validate` → output with outcome `VALIDATED_ONLY`; `Info("success")` with `fileName`, `mode`, counts; table cases: invalid input 400, each file-level 422 code, repository error returned and logged as `Error`, success with rejected rows and warnings (mocks per case, template pattern)
- [x] T063 [US1] DTOs in `backend/domain/dto/asset_import.go` per [contracts/imports-api.md](contracts/imports-api.md): `ImportResultResponse{Outcome, IsCommitted, Mode, FileName, Counts, RejectedRows, Warnings, Import}`, `ImportCountsResponse{Total, Imported, Unchanged, Rejected, WarningRows}`, `ImportRejectedRowResponse{Row, AssetID *string, Errors}`, `ImportIssueResponse{Code, Message}`, `ImportWarningResponse{Row, AssetID, Code, Message}`, `ImportRecordResponse{ID, CreatedAt}`; camelCase JSON
- [x] T064 [US1] Handler package `backend/internal/handler/asset_import/init.go` + `helpers.go` + `helpers_test.go` + `testutil_test.go` (package `asset_import`): `AssetImportHandler` interface; `readUpload(ctx)` (`ctx.FormFile("file")` missing → `400 INVALID_REQUEST`; `mode` form value parsed with `valueobject.ParseImportMode`, empty → `PARTIAL`, invalid → 400); `toImportResultResponse` rendering every issue with `apperr.Message`, `rejectedRows` / `warnings` never `null`, `import` `null` for `VALIDATED_ONLY`, `createdAt` RFC 3339
- [x] T065 [US1] `backend/internal/handler/asset_import/validate_asset_import.go` + `validate_asset_import_test.go`: `POST /api/imports/validate` → `200 "Successfully validate asset import"`; `app.Test` with multipart bodies: success mapping, missing file → 400 with `data.errorCode INVALID_REQUEST`, bad mode → 400, usecase `422` passes through with its `errorCode`, plain error → 500 (mock usecase without `EXPECT()` for parse failures)
- [x] T066 [US1] Register `POST /api/imports/validate` in `backend/internal/routes/router.go` (static route, before any `/:param` route) and wire repositories → usecase → handler in `backend/internal/bootstrap/setup.go` (services → repositories → txManager → usecases → handlers)
- [x] T067 [US1] Usecase integration test `backend/internal/integrationtest/validate_asset_import_test.go`: real repositories + usecase on `TEST_DATABASE_URL`; sample on empty store matches the golden counts and leaves `asset` and `asset_import` empty; a file whose transformer points at a seeded stored substation is accepted (depends on T039, T062)
- [x] T068 [P] [US1] `backend/docs/api/asset-import.md` (English, exception 9): envelope, request parts, `POST /api/imports/validate`, error table incl. `413 IMPORT_FILE_TOO_LARGE`, `IMPORT_FILE_NOT_CSV` with line and `IMPORT_DUPLICATE_COLUMNS`, `ImportResult` fields — matching [contracts/imports-api.md](contracts/imports-api.md)
- [x] T069 [US1] `frontend/src/modules/imports/imports-api.ts` + `imports-api.test.ts` (also create `APIPath.IMPORTS = "/imports"` and `libs/enums/imports-api-path.enum.ts`, deferred from T043 because knip rejects exports nothing uses yet): `validateImport` mutation (`POST APIPath.IMPORTS + ImportsApiPath.VALIDATE`, `FormData` with `file` and `mode`, `extraOptions: { shouldSuppressToast: true }`); export `useValidateImportMutation` from the barrel; test with mocked `fetch` checks the multipart body and the parsed result
- [x] T070 [US1] `frontend/src/pages/assets/components/import-panel/components/import-form/import-form.tsx` + `styles.module.css` + `libs/constants/` + `import-form.test.tsx`: `useAppForm` with `importFileValidationSchema`; labelled file field; mode radio group (Partial — default, Strict) with one-line explanations; **Check file** disabled until a file is chosen, loading + `role="status"` "Checking file…" while pending; changing file or mode calls `onChange` so results reset
- [x] T071 [P] [US1] `frontend/src/pages/assets/components/import-panel/components/outcome-banner/outcome-banner.tsx` + `styles.module.css` + `outcome-banner.test.tsx`: text **and** labelled icon per outcome (never colour alone, SC-005); in this story: "Validated only — nothing saved" and "Not committed — the import failed and was rolled back"
- [x] T072 [P] [US1] `frontend/src/pages/assets/components/import-panel/components/rejected-rows-table/rejected-rows-table.tsx` + `styles.module.css` + `rejected-rows-table.test.tsx`: caption with count; columns Row, Asset ID ("—" + visually hidden "none"), Code, Message; one line per error; sentence "No rows were rejected." when empty
- [x] T073 [P] [US1] `frontend/src/pages/assets/components/import-panel/components/warnings-table/warnings-table.tsx` + `styles.module.css` + `warnings-table.test.tsx`: Row, Asset ID, Code, Message; sentence when empty
- [x] T074 [US1] `frontend/src/pages/assets/components/import-panel/components/import-result/import-result.tsx` + `styles.module.css` + `import-result.test.tsx`: outcome banner, counts list (total, "to import" label for `VALIDATED_ONLY`, unchanged, rejected, rows with warnings), both tables, **Choose another file** (depends on T071–T073)
- [x] T075 [US1] `frontend/src/pages/assets/components/import-panel/import-panel.tsx` + `styles.module.css` + `import-panel.test.tsx`, used inside the `Panel` in `frontend/src/pages/assets/assets.tsx`: owns `useValidateImportMutation`; file-level error shown as `role="alert"` with the server message and no table; result taken from the mutation state (never copied into `useState`); changing file/mode resets the mutation; tests: validate success renders counts and rows, `422` renders the alert, network failure renders the failed outcome, Escape closes and removes `?panel=import`, refresh/deep link shows the empty form (depends on T069, T070, T074)
- [x] T076 [US1] **Checkpoint — real stack**: `docker compose down -v && docker compose up --build`; run [quickstart.md](quickstart.md) API scenarios A, F, G, H, H2, H3 and UI steps 1–4 and 8 (step 2 is the hand check of the `<dialog>` focus trap); confirm `select count(*) from asset` = 0 and `asset_import` is empty after validating; `make check` (with `TEST_DATABASE_URL`), `make lint`, `pnpm lint`, `pnpm build`, `pnpm test` exit 0

---

## Phase 4: User Story 2 — Confirm and save the valid rows (Priority: P1)

**Goal**: After validating, the operator confirms; all valid rows are saved in one atomic step with
an import record, and the result says "Committed — N assets saved".

**Independent Test**: On an empty store, confirm the sample in partial mode → 205 assets (13
substations) with correct parent links, one `asset_import` row `COMMITTED`; a forced database
failure stores nothing and leaves no record.

### Implementation for User Story 2

- [X] T077 [P] [US2] `backend/internal/repository/psql/asset_import/init.go` + `create.go` + `create_test.go` + `testdb_test.go` (package `asset_import_repository`): `Create` inserts via `mapper.ToModelAssetImport`; sqlmock cases: success (`ExpectBegin`/`ExpectExec`/`ExpectCommit` with the full INSERT column list), DB error, `wrapTx`, `unsupportedTx`
- [X] T078 [US2] `backend/internal/repository/psql/asset_import/acquire_lock.go` + `acquire_lock_test.go`: `SELECT pg_advisory_xact_lock(?)` with a fixed `int64` key constant in `init.go`; requires a transaction (`trx == nil` → error, because a transaction-scoped lock outside a transaction is meaningless); sqlmock cases: success, DB error, nil trx, `unsupportedTx` (depends on T077 for `init.go`)
- [x] T079 [P] [US2] `backend/internal/repository/psql/asset/create_many.go` + `create_many_test.go`: `CreateInBatches(models, 1000)`; empty slice → no query; sqlmock cases: success, DB error, `wrapTx`, `unsupportedTx` (depends on T057 for `init.go`)
- [X] T080 [US2] Repository integration test `backend/internal/integrationtest/asset_import_repository_test.go`: `Create` + `CreateMany` with children listed after parents succeed against the real FKs; a second transaction calling `AcquireLock` blocks until the first commits (goroutine + timeout) (depends on T039, T077–T079)
- [X] T081 [US2] **Test first** — `backend/internal/usecase/asset_import/create_asset_import_test.go` (partial mode), seen failing before T082. Table cases with `d.trx` asserted on every repository call and explicit `Commit()` / `Rollback()`: invalid input (no transaction opened); file-level 422 (no transaction opened); `AcquireLock` called first; lock error, find error, record insert error and asset insert error each roll back and are logged; `ToImport > 0` → record `COMMITTED` created **before** `CreateMany`, every inserted asset carries an `ID` from `newID`, the record's `ImportID`, `CreatedAt = UpdatedAt = now()` and `CreatedByName = UpdatedByName = "system"`; `ToImport == 0` → record `NOTHING_TO_COMMIT`, `CreateMany` not called
- [X] T082 [US2] `backend/internal/usecase/asset_import/create_asset_import.go`: log setup → parse (outside the tx) → `WithinTx`: `AcquireLock`, `loadStoredAssets`, `assetimport.Validate`, build `entity.AssetImport` (ID from `newID`, timestamps from `now()`, audit names `usecase.SystemActorName`), `Create` the record, assign `ID` / `ImportID` / audit fields to each accepted asset, `CreateMany`; outcome per data-model.md commit states; `Info("success")` with `importID`, `fileName`, `mode`, `outcome`, counts; technical errors logged as `Error` with `fileName` and `mode` and returned (→ 500, rollback) — T081 passes
- [X] T083 [US2] `backend/internal/handler/asset_import/create_asset_import.go` + `create_asset_import_test.go`: `POST /api/imports` → `201 "Successfully create asset import"` with `import: {id, createdAt}` and `isCommitted = (outcome == COMMITTED)`; multipart parse errors → 400; `500` for plain errors
- [X] T084 [US2] Register `POST /api/imports` in `backend/internal/routes/router.go` (after `/validate`) and add the handler method to the wiring in `backend/internal/bootstrap/setup.go`
- [X] T085 [US2] Usecase integration test `backend/internal/integrationtest/create_asset_import_test.go` (no case here needs the US4 re-import rules): sample on empty store → 205 assets, 13 substations, every `parent_asset_id` resolves, audit names `"system"`, one `COMMITTED` record with counts 222/205/0/17/2, and only `asset` and `asset_import` gained rows (no table holds rejected-row or warning details, FR-031); failure injected mid-transaction (repository wrapper returning an error from `CreateMany`) → no assets, no record (FR-030, SC-006); validate a file whose transformer's parent `SS-X` is missing (`PARENT_NOT_FOUND`), then store `SS-X` with another import, then confirm the first file → the transformer is now committed (FR-022, US2 scenario 3); two concurrent `CreateAssetImport` calls on **non-overlapping** files, where file B's transformers have their parent substation in file A, serialise through the lock; whichever runs second sees the other's committed assets, and the final state is consistent (FR-024) (depends on T039, T082)
- [X] T086 [US2] Large-file test `backend/internal/integrationtest/large_file_test.go` (SC-004, research R6): generate in the test a valid CSV just under 5 MB (~50 000 rows: substations, switchboards, panels, transformers); `ValidateAssetImport` returns within 3 s on the test database; `CreateAssetImport` commits every row, proving the single-array `= ANY(?)` lookups and 1 000-row batches stay under PostgreSQL's bind-parameter limit; skipped without `TEST_DATABASE_URL` and under `go test -short` (depends on T082)
- [X] T087 [P] [US2] Extend `backend/docs/api/asset-import.md` with `POST /api/imports`, the three completed outcomes, `201`/`500` semantics and the import record
- [X] T088 [US2] Add `createImport` mutation (`POST APIPath.IMPORTS`, same `FormData`, `shouldSuppressToast`) to `frontend/src/modules/imports/imports-api.ts` + test in `imports-api.test.ts`; export `useCreateImportMutation`
- [X] T089 [US2] Confirm flow in `frontend/src/pages/assets/components/import-panel/import-panel.tsx` and `components/import-result/import-result.tsx` (+ tests): after a `VALIDATED_ONLY` result, **Save N assets** sends the same `File` and mode with `useCreateImportMutation` (loading "Saving…"); outcome banner texts "Committed — N assets saved" and "Nothing to commit — no new assets" (add to `outcome-banner.tsx` + test); `500` → "Not committed — the import failed and was rolled back"; **Import another file** resets both mutations and the form
- [X] T090 [US2] **Checkpoint — real stack**: `docker compose down -v && docker compose up --build`; quickstart API scenario C and I (stop `db` during a confirm, restart, verify no new record or assets; api log shows the failure with file name and mode) and UI step 5; `psql` shows 205 assets, 13 with type `SUBSTATION`; all five check commands exit 0

---

## Phase 5: User Story 3 — Strict (all-or-nothing) import (Priority: P2)

**Goal**: In strict mode a file with any rejected row stores nothing (but records the attempt);
a fully valid file is saved completely.

**Independent Test**: Confirm the sample in strict mode → 0 assets, one `STRICT_BLOCKED` record,
"Not committed — strict mode, 17 rows rejected".

### Tests for User Story 3 — write first, confirm they FAIL ⚠️

- [X] T091 [US3] Add strict cases to `backend/internal/usecase/asset_import/create_asset_import_test.go`: `STRICT` with ≥ 1 rejected row → record `STRICT_BLOCKED` created, `CreateMany` **not** called, transaction committed, `isCommitted` false; `STRICT` with 0 rejected rows → `COMMITTED`; `STRICT` with 0 rejected and 0 new rows → `NOTHING_TO_COMMIT`
- [X] T092 [P] [US3] Add strict cases to `backend/internal/integrationtest/create_asset_import_test.go`: sample in strict mode → 0 assets, one `STRICT_BLOCKED` record with counts 222/205/0/17/2; a clean 3-row file in strict mode → committed

### Implementation for User Story 3

- [X] T093 [US3] Implement the strict branch in `backend/internal/usecase/asset_import/create_asset_import.go` (checked before the `COMMITTED`/`NOTHING_TO_COMMIT` decision, data-model.md commit states) — T091 and T092 pass
- [X] T094 [US3] Frontend strict outcome: "Not committed — strict mode, N rows rejected" in `frontend/src/pages/assets/components/import-panel/components/outcome-banner/outcome-banner.tsx` (+ test); strict-mode explanation text in `import-form.tsx` enums; `import-result.tsx` keeps listing every rejected row for `STRICT_BLOCKED` (+ test)
- [X] T095 [US3] **Checkpoint — real stack**: `docker compose down -v && docker compose up --build`; quickstart API scenario B and UI step 7; `psql` shows 0 assets and one `STRICT_BLOCKED` record; all five check commands exit 0

---

## Phase 6: User Story 4 — Re-import the same or an overlapping file safely (Priority: P2)

**Goal**: Rows identical to stored assets count as unchanged; rows that reuse a stored ID with
different values are rejected with `ASSET_ID_CONFLICT`; repeating an import adds nothing.

**Independent Test**: Import the sample twice → second result `NOTHING_TO_COMMIT`, 0 imported,
205 unchanged, the same 17 rejections, still 205 assets.

### Tests for User Story 4 — write first, confirm they FAIL ⚠️

- [X] T096 [P] [US4] `backend/domain/entity/asset_test.go` for `DifferingFields(other *Asset) []string`: identical → empty; numbers compared by value (`22` vs `22.0`); dates by calendar day; `nil` vs blank equal; each of the eleven CSV columns detected; result in CSV column order; audit fields, `ID`, `ImportID` ignored
- [X] T097 [P] [US4] `backend/domain/assetimport/reimport_rules_test.go` (FR-018, FR-019): row identical to a stored asset → unchanged (not imported, not rejected, counted in `Unchanged`); same ID with a different status → `ASSET_ID_CONFLICT` naming `operational_status`; several differing fields listed; a child whose in-file parent is rejected for `ASSET_ID_CONFLICT` → `PARENT_REJECTED`; an unchanged row is a valid parent for new children in the same file; sample against a store holding the sample's accepted assets → 0 to import, 205 unchanged, same 17 rejected

### Implementation for User Story 4

- [X] T098 [US4] Implement `DifferingFields` in `backend/domain/entity/asset.go` — T096 passes
- [X] T099 [US4] `backend/domain/assetimport/reimport_rules.go` and wire it into `validate.go` after the hierarchy pass and before `PARENT_REJECTED` propagation, so conflicts propagate to children; unchanged rows are excluded from `Accepted` — T097, T049 and T050 still pass
- [X] T100 [US4] Re-import cases in `backend/internal/integrationtest/create_asset_import_test.go`: sample twice → second `NOTHING_TO_COMMIT`, counts 222/0/205/17, still 205 assets, two records; stored `TX-001-1` with a changed status → `ASSET_ID_CONFLICT` with the field name; a transformer-only file whose parent is stored `SS-001` → committed; **moved from T085**: an asset with the same ID but different values stored between validate and confirm → the confirm result rejects that row with `ASSET_ID_CONFLICT` (US2 scenario 3, conflicting variant); two concurrent imports of **overlapping** files serialise and the second reports the shared rows as unchanged (FR-024, overlapping variant)
- [X] T101 [US4] Frontend re-import rendering tests in `frontend/src/pages/assets/components/import-panel/components/import-result/import-result.test.tsx` and `rejected-rows-table.test.tsx`: unchanged count shown, `ASSET_ID_CONFLICT` message listing fields rendered, "Nothing to commit — no new assets" banner after a re-import response
- [X] T102 [US4] **Checkpoint — real stack**: `docker compose up --build` on the database from T090 (or re-import the sample first); quickstart API scenarios D and E and UI step 6; `psql` shows still 205 assets and a `NOTHING_TO_COMMIT` record; all five check commands exit 0

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Stale template docs, project docs, removal of the template examples (exception 7),
reference-module pointers, and the final Definition of Done run.

- [X] T103 [P] Update stale frontend docs (research R0): `frontend/docs/code-conventions.md` — §16 remove "There is no test runner configured" and add `pnpm test`; §11 error table for the go-common envelope (`data.errorCode`, status fallback) without `UNAUTHENTICATED`/`FORBIDDEN`/`VALIDATION_FAILED`; §12 and §15 without `PrivateRoute`/token text; `frontend/docs/adding-a-feature.md` — register pages under `AppShell`, use `modules/imports` and `pages/assets` as the example
- [X] T104 [P] Update `backend/README.md` and `frontend/README.md` (English, exception 9): prerequisites, run, test commands, no private-module/credentials steps
- [X] T105 [P] `docs/ARCHITECTURE.md`: system overview (compose services), backend layers and the `domain/assetimport` engine, the two-pass order-independent algorithm, validate vs commit endpoints, transaction + advisory lock + outcomes, schema and parent FK on `asset_id`, frontend panel/URL state, key trade-offs (research R2–R7, R10)
- [X] T106 [P] `docs/KNOWN_LIMITATIONS.md`: brief says "220 valid assets" vs 205 accepted (13 substations match); local go-common implements only the used subset (no audit plugin, no telemetry); future commissioned dates and negative voltage accepted (clarification Q2, next steps); detailed results are not stored and no screen lists past import records; 5 MB / ~50 000-row limit; advisory lock serialises imports per database; unconstrained `NUMERIC` compared as `float64`; the template rate limiter (20 non-2xx responses per 15 minutes) can answer `429` during repeated error testing; `backend/values.yml` (Helm) not maintained; template `*_by_name` columns always `"system"`
- [X] T107 [P] Root `README.md`: what the app does, prerequisites, `docker compose up --build`, ports (3000, 8080, 55432), test commands incl. `docker compose --profile test run --rm backend-test` / `frontend-test`, links to `docs/ARCHITECTURE.md`, `docs/KNOWN_LIMITATIONS.md`, `specs/001-csv-import/`, `AGENTS.md`
- [X] T108 Remove the backend `product` example (exception 7): `backend/internal/{handler,usecase,repository/psql}/product/`, `backend/internal/repository/psql/{model,mapper}/product*.go`, `backend/domain/{dto,entity,repository,usecase}/product.go`, `backend/domain/valueobject/{product_status,product_status_test,permission}.go`, `backend/internal/mocks/{repository/ProductRepository.go,usecase/ProductUsecase.go}`, `backend/docs/api/product.md`, product error codes in `domain/errorcode` + `apperr` catalog, and `backend/pkg/gocommon/pkg/middleware/fiber/identity.go`; `make mocks`; the product migration was already removed in T013 (research R8)
- [X] T109 Remove the frontend `notes` example (exception 7): `frontend/src/modules/notes/`, `frontend/src/pages/notes/`, `AppRoute.NOTES`, `APIPath.NOTES`, its route in `frontend/src/index.tsx` and navigation item, and any helper, enum, icon or CSS knip then reports unused
- [X] T110 Point the "reference module" mentions at this feature (depends on T108, T109): root `AGENTS.md` table (`asset_import` backend, `imports` + `pages/assets` frontend); `backend/AGENTS.md` (line "Contoh implementasi referensi: modul `product`" and the `create_product.go` / `find_by_id.go` examples → `asset_import` files); `frontend/AGENTS.md` ("copy the shape of `modules/notes` and `pages/notes`" → `modules/imports` and `pages/assets`). Rule files under `backend/docs/rules/` stay as they are (exception 9)
- [X] T111 [P] Security pass per `backend/docs/rules/07-keamanan.md` checklist (minus auth): no SQL string concatenation, no row contents or secrets in logs, `.env` files untracked (`git ls-files | grep -E '(^|/)\.env$'` empty), errors 500 leak no detail
- [X] T112 Coverage gate: `make cover` in `backend/` reports `internal/usecase` ≥ 80 %; add missing usecase branch tests if not
- [X] T113 Remove the "Sync Impact Report" HTML comment from `.specify/memory/constitution.md` if it is still present (it is review scaffolding, not governance text)
- [X] T114 Final Definition of Done run (constitution): `docker compose down -v && docker compose up --build` starts cleanly; `make check` (with `TEST_DATABASE_URL`) and `make lint` exit 0; `pnpm lint`, `pnpm build`, `pnpm test` exit 0; `docker compose --profile test run --rm backend-test` and `frontend-test` exit 0; full [quickstart.md](quickstart.md) (API A–I incl. H2/H3, UI 1–10, shuffle spot check) passes in the real app; `docs/ARCHITECTURE.md` and `docs/KNOWN_LIMITATIONS.md` reflect the final code; no secrets committed. Record every command with its exit code

---

## Dependencies & Execution Order

### Phase dependencies

- **Setup (Phase 1)**: none. T009 needs T002–T008; T015/T016 need T014; T019 needs T016 + T018; T021 needs all of Phase 1.
- **Foundational (Phase 2)**: needs Phase 1. T026 needs T010; T027 needs T012 + T026; T031 → T032 → T039; T033 → T034; T036 → T037; T038 needs T035 + T037; T042 needs T015, T040, T041.
- **US1 (Phase 3)**: needs Phase 2. Rule tests T044–T050 before T051–T056; T057 before T058; T060 (test) before T061; T063–T064 before T065 → T066.
- **US2 (Phase 4)**: needs US1 (reuses parse/validate helpers, handler package, panel result view). T081 (test) before T082.
- **US3 (Phase 5)**: needs US2 (extends `CreateAssetImport`).
- **US4 (Phase 6)**: needs US2 for the commit path. Without US4, re-importing an already stored ID fails on the unique key, so US4 is required before release; the overlapping/conflicting integration cases live in T100 for that reason.
- **Polish (Phase 7)**: needs all stories. T108/T109 (example removal) come after every story so the references are available throughout. T110 needs T108 + T109. T114 is last.

### User story dependencies

```text
Setup → Foundational → US1 (validate, MVP) → US2 (commit) ─┬→ US3 (strict)
                                                           └→ US4 (re-import)
                                         US3 + US4 → Polish
```

US3 and US4 are independent of each other once US2 is done (US3 touches the outcome branch; US4
touches `domain/entity/asset.go` and `domain/assetimport/reimport_rules.go`; both add cases to
`create_asset_import_test.go` and `integrationtest/create_asset_import_test.go`, so serialise those
two files).

### Within each story

Rule tests → rule implementation → repositories → usecase test → usecase → handler/route →
integration test → frontend → real-stack checkpoint.

## Parallel Examples

```text
# Setup: local go-common packages (different directories)
T003 apperror · T004 response · T005 logger · T006 database · T008 middleware/fiber

# Foundational: value objects, entities, migration 1, interfaces, frontend shell; engine types after T028
T022 · T023 · T024 · T025 · T028 · T029 · T030 · T035 · T040 · T041, then T036

# US1: all rule tests at once, then repositories alongside the engine
T044 · T045 · T046 · T047 · T048 · T049
T057 (asset repo) while T051–T056 (engine) proceed
T071 · T072 · T073 (result sub-components) · T068 (API doc)

# US2: repository methods
T077 · T079, then T078

# US4: both test files
T096 · T097

# Polish: docs
T103 · T104 · T105 · T106 · T107 · T111
```

## Implementation Strategy

### MVP first (User Story 1)

1. Phase 1 Setup → T021 checkpoint (stack runs, all checks green).
2. Phase 2 Foundational → schema migrates on start, panel opens via URL.
3. Phase 3 US1 → validation with every rule, golden + permutation tests, panel shows results.
4. **Stop and validate** at T076 in the real stack: it is a usable data-quality checker that stores nothing.

### Incremental delivery

- - US2 (T077–T090): imports can be committed atomically → the core value of the brief.
- - US3 (T091–T095): strict mode.
- - US4 (T096–T102): safe re-import; required before release (see dependencies).
- Polish (T103–T114): docs, example removal, final Definition of Done.

## Notes

- `[P]` = different files and no dependency on an unfinished task.
- Rule tests must be seen failing before their implementation task starts (Principle IV).
- Each task leaves `make check` / `pnpm test` green unless it is a test-first task whose purpose is
  to fail; commit after each task or logical group (Conventional Commits).
- Never edit generated mocks; run `make mocks` after changing a `domain/` interface.
- Migrations 000001/000002 are never edited once merged; schema changes need a new migration.
