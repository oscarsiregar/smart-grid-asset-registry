# 01 — Architecture

## Principle

This service uses a simple **Clean Architecture** with three working layers
(handler → usecase → repository/service) and one contract layer (`domain/`).
The goal: business logic can be tested without HTTP and without a database, and each
layer can be replaced without touching the others.

```
                 ┌──────────────────────────── domain/ (contracts, no framework) ────────────────────────────┐
                 │ entity · valueobject · errorcode · dto · repository(interface) · service(interface) · usecase(interface + Input/Output) │
                 └───────────────▲─────────────────────────▲───────────────────────────▲─────────────────────┘
                                 │                         │                           │
HTTP ─► routes ─► internal/handler ─► internal/usecase ─► internal/repository/psql ─► PostgreSQL
                    (parse & map)     (business rules) └► internal/service/*       ─► external APIs
```

## Folder structure

```
cmd/api/main.go                  entry point: config, DB, Fiber, middleware, graceful shutdown
domain/
  entity/                        plain domain structs (no gorm/json tags)
  valueobject/                   enums (Parse/Scan/Value/String/IsValid) + permission constants
  errorcode/                     error code constants (+ the All list)
  dto/                           HTTP contract: Request (query/params/json tags) & Response (camelCase json)
  repository/                    repository interfaces + Tx/TransactionManager + filter structs
  service/                       interfaces for outbound integrations (other APIs, storage, brokers)
  usecase/                       usecase interfaces + XxxInput/XxxOutput + Validate()/SetDefault()
internal/
  apperr/                        error message catalog (ID + EN) → go-common apperror
  bootstrap/                     composition root (DI) + Fiber configuration
  config/                        loading & validating env
  handler/<module>/              HTTP handlers, 1 file = 1 endpoint
  middleware/                    local middleware (recover, limiter)
  mocks/                         mockery output (+ hand-written txhelpers.go)
  repository/psql/
    gormtx/                      GORM transaction adapter ↔ domain Tx
    model/                       GORM models (tags + TableName)
    mapper/                      model ↔ entity
    querysafety/                 ORDER BY & LIKE sanitising
    <module>/                    repository implementation, 1 file = 1 method
  routes/router.go               the only place routes are registered
  service/<name>/                domain/service implementation, 1 file = 1 call
  usecase/<module>/              business logic, 1 file = 1 use case
migrations/                      ordered SQL migrations
docs/                            rules, guides, API contracts
```

## Dependency rules (MUST)

| Package | May import | Must not import |
|---|---|---|
| `domain/*` | stdlib, `uuid`, other `domain/*` | `internal/*`, gorm, fiber, go-common, viper |
| `internal/handler` | `domain/*`, `internal/apperr`, fiber, go-common (response, middleware) | `internal/repository`, `internal/usecase/<impl>`, gorm |
| `internal/usecase` | `domain/*`, `internal/apperr`, go-common `logger`, logrus | fiber, gorm, `internal/repository`, `internal/handler` |
| `internal/repository/psql` | `domain/*`, gorm, other `psql/*` | `internal/apperr`, fiber, `internal/usecase` |
| `internal/service` | `domain/*`, net/http, go-common `utils` | `internal/repository`, `internal/usecase` |
| `internal/bootstrap` | everything (composition root) | — |

> Why is `domain/` kept clean? Because its interfaces & types are used by every layer.
> If `domain` imported gorm/fiber, a framework change would spread everywhere and
> mocks could not be generated without pulling in every dependency.

## Responsibilities per layer

| Layer | Does | Does NOT |
|---|---|---|
| Handler | parse query/params/body, convert types (UUID, enum), read the caller's identity, call 1 usecase, map Output → DTO, choose the success status | check permissions, business rules, DB queries, transactions, format error responses |
| Usecase | permissions, validation, business rules, orchestrating repos/services, transactions, logging | HTTP parsing, SQL, JSON formatting |
| Repository | 1 data operation, model↔entity mapping, propagating technical errors | business rules, `apperr`, calling other repositories, opening its own transaction |
| Service (outbound) | call external APIs, map responses to the domain Output, timeouts | business rules |

## Composition root

`internal/bootstrap/setup.go` is the only place where concrete `NewXxx(...)` constructors
are called. Order: **services → repositories → txManager → usecases → handlers**,
then returned as `routes.RouteDependencies`. There are no global variables
or `init()` functions holding dependencies.

## When to create a new module vs extend one

- **A new module** (`internal/{handler,usecase,repository/psql}/<module>`) for every
  main aggregate/table that has its own endpoints.
- An endpoint that changes data owned by another aggregate stays in the module that
  **owns the route** (e.g. `/periods/:id/…` in a `period_*` module), and its usecase may use
  several repositories.
- A usecase does **not** call another usecase. Shared logic → a function in that module's
  `helpers.go`, or a method on the entity when it is pure domain logic.
