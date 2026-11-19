# Smart Grid Asset Registry — backend

Go API that validates grid-asset CSV files, stores the accepted assets in PostgreSQL, and serves the stored
hierarchy to the explorer.
**Go 1.26 · Fiber v2 · GORM / PostgreSQL · local go-common**, layered handler → usecase → repository with the
contracts in `domain/`.

> **Rules: [AGENTS.md](AGENTS.md)** (short, binding) · details in **[docs/rules/](docs/rules/)** (Indonesian,
> kept from the template) · adding a feature: **[docs/menambah-fitur.md](docs/menambah-fitur.md)** · API:
> **[docs/api/asset-import.md](docs/api/asset-import.md)**, **[docs/api/asset.md](docs/api/asset.md)** · architecture: [../docs/ARCHITECTURE.md](../docs/ARCHITECTURE.md)

## Prerequisites

- Go 1.26
- PostgreSQL 16 — the easiest is the compose `db` service: `docker compose up -d db` from the repository root
  (host port `55432`)
- Optional: `golangci-lint` v2.13 for `make lint`, `mockery` v2.53 for `make mocks`

No private modules and no credentials are needed: the template's private `go-common` library is re-implemented
under `pkg/gocommon/` (constitution, Template Exception 4).

## Run

```sh
cp .env.example .env    # local values that match the compose db; never commit .env
make run                # http://localhost:8080/api/health
```

Migrations in `migrations/` are embedded in the binary and applied on start (golang-migrate). There is no
authentication: every route is public and lives under `/api` (Template Exceptions 1 and 5).

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/api/health` | liveness / readiness (pings the database) |
| `POST` | `/api/imports/validate` | check a CSV file against the stored assets; stores nothing |
| `POST` | `/api/imports` | re-validate and commit a CSV file in one transaction |
| `GET` | `/api/assets[?parentAssetId=]` | the top level (every substation) or one asset's children, complete and ordered |
| `GET` | `/api/assets/:assetId` | one asset (ID percent-encoded): attributes, path, children, and counts for a substation |

## Test

```sh
make check   # gofmt + vet + tests
make lint    # golangci-lint v2
make cover   # internal/usecase coverage (gate >= 80 %)
```

The integration tests in `internal/integrationtest/` run against a real PostgreSQL database and are **skipped**
without `TEST_DATABASE_URL` (Template Exception 11). With the compose stack running, the `asset_registry_test`
database already exists:

```sh
TEST_DATABASE_URL='postgres://postgres:postgres@localhost:55432/asset_registry_test?sslmode=disable' make check
```

Or run the same check in a container from the repository root: `docker compose --profile test run --rm backend-test`.

## Commands

| Command | Purpose |
|---|---|
| `make run` | run the API (reads `.env`) |
| `make check` | gofmt + vet + test — run before every push |
| `make test` | tests (same package list as CI) |
| `make cover` | `internal/usecase` coverage (CI gate ≥ 80 %) |
| `make lint` | golangci-lint v2 |
| `make mocks` | regenerate mocks from the `domain/` interfaces |
| `make build` | binary in `bin/api` |

## Layout

```
cmd/api/                  entry point (config, DB, migrations, middleware, graceful shutdown)
domain/                   framework-free contracts: entity, valueobject, errorcode, dto,
                          repository/usecase interfaces
domain/assetimport/       the pure import rules: parse + validate a CSV against stored assets
                          (standard library and domain/* only, Template Exception 10)
internal/
  apperr/                 English error-message catalog
  bootstrap/              dependency injection + Fiber configuration
  config/                 load and validate env
  handler/asset_import/   HTTP handlers — one file per endpoint
  integrationtest/        tests against real PostgreSQL (TEST_DATABASE_URL)
  middleware/             recover, rate limiter
  mocks/                  mockery output
  repository/psql/        gormtx, model, mapper, querysafety, asset/, asset_import/ (one file per method)
  routes/                 route registration
  testsupport/pgtest/     opens, migrates and truncates the test database
  usecase/asset_import/   business flow — one file per use case
migrations/               ordered SQL migrations, embedded
pkg/gocommon/             local re-implementation of go-common
docs/api/                 API contract per module
docs/rules/               development rules
```

Request flow:

```
client ─► [Recover → CommonMiddleware → Limiter → Compress]
       ─► /api group ─► handler (parse multipart) ─► usecase (validate, rules, transaction + advisory lock)
       ─► repository (GORM) ─► PostgreSQL
       ◄─ response.Success / global ErrorHandler (AppError → status, message, data.errorCode)
```

## Rule documents

| # | Topic |
|---|---|
| [01](docs/rules/01-arsitektur.md) | Architecture and dependency direction |
| [02](docs/rules/02-penamaan.md) | Naming and file layout |
| [03](docs/rules/03-handler.md) | HTTP handlers |
| [04](docs/rules/04-usecase.md) | Usecases, validation, transactions |
| [05](docs/rules/05-repository-database.md) | Repositories, GORM, migrations |
| [06](docs/rules/06-error-handling.md) | Error handling |
| [07](docs/rules/07-keamanan.md) | Security |
| [08](docs/rules/08-testing.md) | Testing |
| [09](docs/rules/09-logging-observability.md) | Logging, health check, shutdown |
| [10](docs/rules/10-git-ci-deploy.md) | Git, CI, deploy |

The rule documents are in Indonesian and describe the company template, including authentication; where this
project differs, the constitution's Template Exceptions apply (`../.specify/memory/constitution.md`).
