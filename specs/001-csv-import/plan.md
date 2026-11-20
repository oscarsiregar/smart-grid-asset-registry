# Implementation Plan: CSV Import of Grid Assets

**Branch**: `001-csv-import` | **Date**: 2026-11-19 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/001-csv-import/spec.md`

## Summary

Operators upload a grid-asset CSV, see every accepted, unchanged and rejected row with stable error
codes, and then confirm a partial (default) or strict import that is applied atomically and
serialised by a PostgreSQL advisory lock. The import rules live in a pure Go package
(`domain/assetimport`) that receives the relevant stored assets as a map, so every rule is
table-tested and the sample file is pinned by a golden test. The backend follows the template's
layers (handler → usecase → repository) for two new modules, `asset_import` and `asset`. The
frontend adds an import panel over a new `/assets` page, with its open state in the URL. The work
also delivers the project foundation the constitution requires: module rename, a local go-common,
embedded migrations run on start, auth removal, Vitest, and a root compose stack with a test profile.

## Technical Context

**Language/Version**: Go 1.26 (backend); TypeScript 6 / React 19 on Node 24 (frontend)

**Primary Dependencies**: Fiber v2, GORM + pgx v5, logrus, golang-migrate v4 (new: `iofs` source +
`pgx/v5` driver), local go-common under `backend/pkg/gocommon`; Vite 8, RTK Query, React Router 7,
React Hook Form + Zod; new dev deps Vitest, jsdom, Testing Library (react, user-event, jest-dom)

**Storage**: PostgreSQL 16 — tables `asset_import`, `asset` ([data-model.md](data-model.md))

**Testing**: Go: testify + mockery (usecase), sqlmock (repository), table-driven unit tests and a
golden test for `domain/assetimport`, real-PostgreSQL integration tests via `TEST_DATABASE_URL`
(skipped when unset). Frontend: Vitest + React Testing Library with mocked `fetch`.

**Target Platform**: Linux containers (compose: `db`, `api`, `web` nginx); evergreen browsers

**Project Type**: Web application (Go API + React SPA)

**Performance Goals**: Validation result within 3 s for the sample (SC-004); the engine is linear in
rows plus two indexed lookups

**Constraints**: File ≤ 5 MB (spec Assumptions); one import at a time (FR-024); Fiber `BodyLimit`
6 MB; no CHECK constraints or triggers (Constitution II); messages in English (exception 6)

**Scale/Scope**: ~50 000 rows per file at the 5 MB limit; registry size unbounded (lookups are
by the IDs in the file, never full-table)

No open NEEDS CLARIFICATION items; all decisions are in [research.md](research.md).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|---|---|---|
| I. Follow the template | Pass | New code copies the `product` (backend) and `notes` (frontend) shapes. Every deviation is a listed exception: the five accepted in 1.1.0 are exceptions 8–12. The rejected one (D2) is dropped, so both tables keep the template audit columns with `"system"` as the actor name ([research.md R0, R7](research.md#r0-template-fit-deviations-that-need-a-decision)) |
| II. Validation lives in the backend | Pass | All rules in `domain/assetimport` + usecases; schema uses PK/unique/NOT NULL/FK only, no CHECK (exception 8); the frontend Zod schema checks only "a file is chosen" and the mode value |
| III. Predictable, explainable imports | Pass | Set/map-based two-pass engine (R3), stable codes in `domain/errorcode`, row number + asset ID on every rejection, `outcome` + `isCommitted` in every result, `500` ⇒ rolled back |
| IV. Test the rules | Pass | Table test per rule written first; golden test on the sample (R12); permutation test; repository and usecase integration tests on real PostgreSQL (exception 11) |
| V. Usable and accessible | Pass | `?panel=import` in the URL, native modal `<dialog>` (focus trap, Escape), loading/empty/error states listed in [contracts/ui-import-panel.md](contracts/ui-import-panel.md), outcomes in text plus labelled icons |
| Template exceptions used | 1–12 | auth removal, Vitest, compose + migrate on start, local go-common (+`data.errorCode`), `/api`, English catalog, example removal, no CHECK constraints, English docs, `domain/assetimport`, integration-test packages, in-memory uploads |
| Definition of Done | Planned | `make check`, `make lint`, `pnpm lint/build/test`, `docker compose up --build` + manual run ([quickstart.md](quickstart.md)), `docs/ARCHITECTURE.md` + `docs/KNOWN_LIMITATIONS.md` created, no secrets (`.env` files stay git-ignored) |

**Post-design re-check (constitution 1.1.0, 2026-11-19)**: **PASS**, no open violations. D1, D3,
D4, D5 and D6 are now exceptions 8–12. D2 was rejected and is designed out: `asset_import` and
`asset` carry `created_by_name` / `updated_by_name` set to `usecase.SystemActorName` (`"system"`)
([data-model.md](data-model.md)). The go-common provenance note in research R1 is confirmed
acceptable (signatures and behaviour only, no copying).

## Project Structure

### Documentation (this feature)

```text
specs/001-csv-import/
├── plan.md                    # this file
├── research.md                # decisions, template-fit deviations (R0), provisional golden values
├── data-model.md              # tables, domain types, rule → code table, commit states
├── quickstart.md              # end-to-end validation guide
├── contracts/
│   ├── imports-api.md         # HTTP contract
│   └── ui-import-panel.md     # URL state, panel behaviour, UI states
└── tasks.md                   # /speckit-tasks (not created here)
```

### Source Code (repository root)

```text
.gitattributes                                   # LF everywhere, binaries marked
docker-compose.yml                               # db (5432), api (8080), web (3000); profile "test"
docker/postgres/init/01-create-test-db.sql       # creates asset_registry_test
docs/ARCHITECTURE.md                             # layers, import flow, transaction, trade-offs
docs/KNOWN_LIMITATIONS.md

backend/
├── go.mod                                       # module smart-grid-asset-registry (+ golang-migrate)
├── Makefile, .mockery.yaml, .env.example, Dockerfile, .gitlab-ci.yml   # module name, no private module
├── cmd/api/main.go                              # + run embedded migrations before Fiber starts
├── migrations/
│   ├── migrations.go                            # //go:embed *.sql
│   ├── 000001_create_asset_import.{up,down}.sql
│   └── 000002_create_asset.{up,down}.sql        # (000001_create_product.* deleted)
├── pkg/gocommon/                                # local go-common (exception 4), upstream paths kept
│   ├── pkg/apperror/  pkg/logger/  pkg/database/  pkg/utils/  pkg/middleware/fiber/
│   └── response/
├── domain/
│   ├── assetimport/                             # pure engine (exc. 10): parse, normalize, field_rules,
│   │   ├── *.go + *_test.go                     #   hierarchy_rules, reimport_rules, serial_warnings,
│   │   └── testdata/grid_assets.csv, grid_assets.golden.json   # validate, types
│   ├── entity/{asset.go, asset_import.go}
│   ├── valueobject/{asset_type.go, operational_status.go, import_mode.go, import_outcome.go} (+ tests)
│   ├── errorcode/errorcode.go                   # + file, row and warning codes
│   ├── dto/asset_import.go                      # CreateAssetImportRequest, ImportResultResponse, …
│   ├── repository/{asset.go, asset_import.go}   # interfaces
│   └── usecase/asset_import.go                  # AssetImportUsecase + Input/Output + Validate()
├── internal/
│   ├── apperr/apperr.go                         # + Message(), ErrorCode on AppError, EN only
│   ├── config/config.go                         # drop JWT/internal secret; APP_PORT 8080 default
│   ├── bootstrap/{setup.go, fiber.go}           # wiring; BodyLimit 6 MB
│   ├── routes/router.go                         # /api group, /api/health, /api/imports[/validate]
│   ├── handler/asset_import/                    # init.go, validate_asset_import.go, create_asset_import.go, helpers.go (+ tests)
│   ├── usecase/asset_import/                    # init.go, validate_asset_import.go, create_asset_import.go, helpers.go (+ tests)
│   ├── repository/psql/
│   │   ├── model/{asset.go, asset_import.go}
│   │   ├── mapper/{asset.go, asset_import.go} (+ tests)
│   │   ├── asset/                               # init, find_by_asset_ids, find_by_serial_numbers, create_many (+ sqlmock tests)
│   │   └── asset_import/                        # init, create, acquire_lock (+ sqlmock tests)
│   ├── mocks/                                   # regenerated (make mocks)
│   ├── testsupport/pgtest/                      # open TEST_DATABASE_URL, migrate, truncate, skip (exc. 11)
│   └── integrationtest/                         # repositories + usecases on real PostgreSQL (exc. 11)
└── docs/api/asset-import.md                     # contract (English, D3)

frontend/
├── package.json, pnpm-lock.yaml                 # + test script and dev deps (exception 2)
├── vite.config.ts, knip.config.ts, .env.example # Vitest block; VITE_APP_API_ORIGIN_URL=/api
├── nginx/nginx.conf                             # upstream api:8080
├── docs/{code-conventions.md, adding-a-feature.md}   # remove auth/"no test runner" text
└── src/
    ├── index.tsx                                # routes without PrivateRoute; / → /assets
    ├── libs/
    │   ├── components/app-shell/                # was authenticated-shell, no account area
    │   ├── components/panel/                    # native <dialog> modal panel (+ test)
    │   ├── enums/{app-route, api-path, error-code, search-parameter, panel-name}.enum.ts
    │   ├── hooks/use-panel/use-panel.hook.ts    # read/open/close ?panel= (+ test)
    │   ├── modules/api/libs/helpers/{base-query, to-server-error}.helper.ts   # envelope + no auth (+ tests)
    │   └── modules/testing/                     # Vitest setup
    ├── modules/imports/
    │   ├── imports-api.ts, imports.ts
    │   └── libs/{enums, types, validation-schemas}/
    └── pages/assets/
        ├── assets.tsx, styles.module.css
        ├── components/import-panel/
        │   ├── import-panel.tsx
        │   └── components/{import-form, import-result, outcome-banner, rejected-rows-table, warnings-table}/
        └── libs/{enums, constants, helpers}/
```

Removed by this feature: backend `product` module (handler, usecase, repository, model, mapper,
DTO, entity, valueobject, mocks, docs) and `domain/valueobject/permission.go`; frontend
`modules/{auth,users,notes}`, `pages/{auth,dashboard,no-access,notes}`, `private-route`, token
storage helpers (exception 7). The examples stay in place as the pattern reference until the new
modules exist, then go in the final cleanup phase. Their migration goes in the foundation phase
(R8).

**Structure Decision**: Web application with the template's existing `backend/` and `frontend/`
trees. New code goes into the template's own folders, plus the two backend locations that
exceptions 10 and 11 allow: `domain/assetimport` and `internal/{testsupport/pgtest,integrationtest}`.

## Implementation phases (input for `/speckit-tasks`)

1. **Foundation**: `.gitattributes`; module rename; local go-common (+ tests for the `errorCode`
   envelope); config without secrets; `/api` router without the auth chain; delete the product
   migration and routes; embedded migrations + golang-migrate on start; compose stack + test DB;
   frontend auth removal, `AppShell`, `toServerError`/`baseQuery` for the envelope, Vitest setup.
   Exit: `docker compose up --build` starts; `/api/health` is 200; all checks are green.
2. **Domain + pure engine (test-first)**: value objects; entities (`DifferingFields`); error codes +
   catalog; `assetimport` rule by rule, each table test written before its rule; golden test frozen
   after the hand check against research R12; permutation test.
3. **Persistence**: migrations 000001/000002; models, mappers, repositories with sqlmock tests;
   `pgtest` helper; repository integration tests.
4. **Usecases + handlers**: `ValidateAssetImport`, `CreateAssetImport` (lock, tx, outcomes,
   logging) with mocked tests (≥ 80 % coverage) and integration tests (commit, strict, re-import,
   conflict, rollback on failure, two concurrent imports); handlers + `backend/docs/api/asset-import.md`.
5. **Frontend**: `panel` component + `use-panel` hook; `modules/imports`; `pages/assets` with the
   import panel and all states from the UI contract; component tests (states, URL, Escape, outcomes).
6. **Cleanup + docs**: remove template examples (exception 7) and stale doc text; `docs/ARCHITECTURE.md`,
   `docs/KNOWN_LIMITATIONS.md`; run the full [quickstart.md](quickstart.md).

## Complexity Tracking

No open violations. The deviations proposed during planning were resolved by constitution 1.1.0:

| Proposal | Outcome |
|---|---|
| D1 No CHECK constraints for enums/ranges | Accepted — exception 8 |
| D2 Drop `*_by_name` audit columns | **Rejected** — columns kept, filled with `"system"` |
| D3 Project docs in English | Accepted — exception 9 |
| D4 Package `domain/assetimport` | Accepted — exception 10 |
| D5 Packages `internal/integrationtest`, `internal/testsupport/pgtest` | Accepted — exception 11 |
| D6 CSV parsed in memory, never stored | Accepted — exception 12 |
