# Unit test report

Smart Grid Asset Registry — generated 2026-11-20 14:26 SE Asia Standard Time on branch `005-ui-polish`, on top of commit `742fa8e` with the branch's uncommitted changes.

## Summary

| Suite | Command | Tests passed | Failed | Skipped | Exit code |
|---|---|---|---|---|---|
| Backend unit tests | `go test ./...` (in `backend/`) | 545 | 0 | 25 | 0 |
| Backend integration tests (PostgreSQL 16) | `TEST_DATABASE_URL=… go test ./internal/integrationtest/... ./internal/testsupport/...` | 66 | 0 | 1 | 0 |
| Frontend unit and component tests | `pnpm test` (Vitest, in `frontend/`) | 259 | 0 | 0 | 0 |

Counts include sub-tests (`t.Run`) for Go and every `it` for Vitest. In the unit run the integration tests skip themselves without `TEST_DATABASE_URL` (Template Exception 11); they run in the second row.

**Coverage gate** (`make cover`, `internal/usecase`, required ≥ 80 %): **99.4 %** — pass.

## Environment

- Go: `go version go1.26.11 windows/amd64`
- Node.js: `v24.19.0`, Vitest 5.1.0, React Testing Library, jsdom
- PostgreSQL 16 (the `db` service of `docker-compose.yml`, database `asset_registry_test`) for the integration tests

## Backend — per package

Coverage is the share of the package's own statements covered by its tests (`go test -cover`).

| Package | Passed | Failed | Skipped | Coverage |
|---|---|---|---|---|
| `domain/assetimport` | 88 | 0 | 0 | 97.9 % |
| `domain/entity` | 55 | 0 | 0 | 100.0 % |
| `domain/usecase` | 32 | 0 | 0 | 100.0 % |
| `domain/valueobject` | 64 | 0 | 0 | 100.0 % |
| `internal/apperr` | 10 | 0 | 0 | 92.9 % |
| `internal/config` | 8 | 0 | 0 | 46.4 % |
| `internal/handler/asset` | 33 | 0 | 0 | 98.3 % |
| `internal/handler/asset_import` | 17 | 0 | 0 | 97.7 % |
| `internal/integrationtest` | 0 | 0 | 23 | — |
| `internal/middleware` | 11 | 0 | 0 | 100.0 % |
| `internal/repository/psql/asset` | 61 | 0 | 0 | 100.0 % |
| `internal/repository/psql/asset_import` | 10 | 0 | 0 | 100.0 % |
| `internal/repository/psql/gormtx` | 7 | 0 | 0 | 100.0 % |
| `internal/repository/psql/mapper` | 9 | 0 | 0 | 100.0 % |
| `internal/repository/psql/querysafety` | 11 | 0 | 0 | 100.0 % |
| `internal/testsupport/pgtest` | 1 | 0 | 2 | 14.8 % |
| `internal/usecase/asset` | 43 | 0 | 0 | 98.7 % |
| `internal/usecase/asset_import` | 44 | 0 | 0 | 100.0 % |
| `migrations` | 7 | 0 | 0 | 45.0 % |
| `pkg/gocommon/pkg/apperror` | 7 | 0 | 0 | 100.0 % |
| `pkg/gocommon/pkg/logger` | 6 | 0 | 0 | 100.0 % |
| `pkg/gocommon/pkg/middleware/fiber` | 3 | 0 | 0 | 94.7 % |
| `pkg/gocommon/pkg/utils` | 12 | 0 | 0 | 83.3 % |
| `pkg/gocommon/response` | 6 | 0 | 0 | 100.0 % |

Packages without tests — wiring (`cmd/api`, `internal/bootstrap`, `internal/routes`), declarations (DTOs, interfaces, error codes, GORM models) or a thin driver wrapper; their code is exercised by the handler, usecase and integration tests: `cmd/api`, `domain/dto`, `domain/errorcode`, `domain/repository`, `internal/bootstrap`, `internal/repository/psql/model`, `internal/routes`, `pkg/gocommon/pkg/database`.

`internal/integrationtest` and the real-database tests of `internal/testsupport/pgtest` show as skipped here because `TEST_DATABASE_URL` is unset in this run; the next section runs them.

## Backend — integration tests on PostgreSQL

| Package | Passed | Failed | Skipped |
|---|---|---|---|
| `internal/integrationtest` | 64 | 0 | 0 |
| `internal/testsupport/pgtest` | 2 | 0 | 1 |

The one skipped test is `TestOpen_SkipsWithoutDatabaseURL`: it clears `TEST_DATABASE_URL` on purpose and checks that a real-database test then skips itself.

## Frontend — per test file

| Test file | Passed | Failed | Skipped |
|---|---|---|---|
| `src/libs/components/button/button.test.tsx` | 2 | 0 | 0 |
| `src/libs/components/logo/logo.test.tsx` | 2 | 0 | 0 |
| `src/libs/components/panel/panel.test.tsx` | 7 | 0 | 0 |
| `src/libs/components/skeleton/skeleton.test.tsx` | 4 | 0 | 0 |
| `src/libs/hooks/use-panel/use-panel.hook.test.tsx` | 6 | 0 | 0 |
| `src/libs/modules/api/libs/helpers/to-server-error.helper.test.ts` | 8 | 0 | 0 |
| `src/libs/modules/api/libs/helpers/unwrap-envelope.helper.test.ts` | 6 | 0 | 0 |
| `src/modules/assets/assets-api.test.tsx` | 7 | 0 | 0 |
| `src/modules/assets/assets.test.ts` | 3 | 0 | 0 |
| `src/modules/assets/libs/validation-schemas/asset-search.validation-schema.test.ts` | 11 | 0 | 0 |
| `src/modules/imports/imports-api.test.tsx` | 10 | 0 | 0 |
| `src/modules/imports/imports.test.ts` | 5 | 0 | 0 |
| `src/modules/imports/libs/helpers/create-import-form-data.helper.test.ts` | 1 | 0 | 0 |
| `src/pages/assets/assets.test.tsx` | 29 | 0 | 0 |
| `src/pages/assets/components/asset-details/asset-details.test.tsx` | 10 | 0 | 0 |
| `src/pages/assets/components/asset-details/components/child-groups/child-groups.test.tsx` | 4 | 0 | 0 |
| `src/pages/assets/components/asset-details/components/child-list/child-list.test.tsx` | 2 | 0 | 0 |
| `src/pages/assets/components/asset-details/components/descendant-counts/descendant-counts.test.tsx` | 2 | 0 | 0 |
| `src/pages/assets/components/asset-details/components/details-skeleton/details-skeleton.test.tsx` | 1 | 0 | 0 |
| `src/pages/assets/components/asset-header/asset-header.test.tsx` | 3 | 0 | 0 |
| `src/pages/assets/components/asset-search/components/search-form/components/asset-type-field/asset-type-field.test.tsx` | 3 | 0 | 0 |
| `src/pages/assets/components/asset-search/components/search-form/search-form.test.tsx` | 10 | 0 | 0 |
| `src/pages/assets/components/asset-search/components/search-results/search-results.test.tsx` | 12 | 0 | 0 |
| `src/pages/assets/components/asset-tree/asset-tree.test.tsx` | 8 | 0 | 0 |
| `src/pages/assets/components/asset-tree/components/tree-skeleton/tree-skeleton.test.tsx` | 1 | 0 | 0 |
| `src/pages/assets/components/import-panel/components/import-form/import-form.test.tsx` | 5 | 0 | 0 |
| `src/pages/assets/components/import-panel/components/import-result-actions/import-result-actions.test.tsx` | 4 | 0 | 0 |
| `src/pages/assets/components/import-panel/components/import-result/import-result.test.tsx` | 5 | 0 | 0 |
| `src/pages/assets/components/import-panel/components/import-steps/import-steps.test.tsx` | 3 | 0 | 0 |
| `src/pages/assets/components/import-panel/components/outcome-banner/outcome-banner.test.tsx` | 4 | 0 | 0 |
| `src/pages/assets/components/import-panel/components/rejected-rows-table/rejected-rows-table.test.tsx` | 5 | 0 | 0 |
| `src/pages/assets/components/import-panel/components/warnings-table/warnings-table.test.tsx` | 2 | 0 | 0 |
| `src/pages/assets/components/import-panel/import-panel.test.tsx` | 13 | 0 | 0 |
| `src/pages/assets/components/import-panel/libs/helpers/check-is-failure.helper.test.ts` | 4 | 0 | 0 |
| `src/pages/assets/components/import-panel/libs/helpers/get-import-step.helper.test.ts` | 5 | 0 | 0 |
| `src/pages/assets/components/import-panel/libs/helpers/get-outcome-summary.helper.test.ts` | 4 | 0 | 0 |
| `src/pages/assets/components/register-summary/register-summary.test.tsx` | 4 | 0 | 0 |
| `src/pages/assets/libs/helpers/check-is-addressable-asset-id.helper.test.ts` | 5 | 0 | 0 |
| `src/pages/assets/libs/helpers/check-is-asset-not-found.helper.test.ts` | 5 | 0 | 0 |
| `src/pages/assets/libs/helpers/get-asset-location-text.helper.test.ts` | 2 | 0 | 0 |
| `src/pages/assets/libs/helpers/get-asset-path.helper.test.ts` | 4 | 0 | 0 |
| `src/pages/assets/libs/helpers/get-search-summary-text.helper.test.ts` | 6 | 0 | 0 |
| `src/pages/assets/libs/hooks/use-asset-search/use-asset-search.hook.test.tsx` | 7 | 0 | 0 |
| `src/pages/assets/libs/hooks/use-expanded-nodes/use-expanded-nodes.hook.test.tsx` | 6 | 0 | 0 |
| `src/pages/assets/libs/hooks/use-tree-keyboard/use-tree-keyboard.hook.test.tsx` | 7 | 0 | 0 |
| `src/pages/assets/libs/maps.test.ts` | 2 | 0 | 0 |

## How to reproduce

```sh
docker compose up -d db                                   # PostgreSQL for the integration tests
cd backend
go test ./... -cover                                      # unit tests, per-package coverage
make cover                                                # internal/usecase coverage gate (>= 80 %)
TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5432/asset_registry_test?sslmode=disable' \
  go test ./internal/integrationtest/... ./internal/testsupport/...
cd ../frontend && pnpm test                               # Vitest
```

In containers: `docker compose --profile test run --rm backend-test` and `docker compose --profile test run --rm frontend-test`.
