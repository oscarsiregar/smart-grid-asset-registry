# AGENTS.md — backend rules (MUST be read before writing code)

This file applies to humans **and** AI agents (Claude Code, Codex, Cursor, etc.).
The summary here is **binding**; the details & reasons are in [docs/rules/](docs/rules/).
Reference implementation: the **`asset_import`** module (`internal/handler/asset_import`, `internal/usecase/asset_import`,
`internal/repository/psql/asset` and `asset_import`; pure rules in `domain/assetimport`) — copy its patterns, don't invent new ones.
The template's example module `product` has been removed (constitution → *Template Exceptions* no. 7); where [docs/adding-a-feature.md](docs/adding-a-feature.md)
or the rules mention `product` files, their counterparts are in the `asset_import` module (e.g. `create_product.go` → `create_asset_import.go`).

> **This project:** there is no authentication/authorisation (an approved exception to the template, see
> `.specify/memory/constitution.md` → *Template Exceptions* no. 1). Business routes are public under `/api`; there is no
> user identity, permission or JWT middleware. The other rules apply unchanged.

## Stack
Go 1.26 · Fiber v2 · GORM + PostgreSQL · `go-common` (response, apperror, logger, middleware, database) · testify + mockery v2 · sqlmock.

> **This project:** `go-common` is the company's private library and is **re-implemented locally** inside
> `backend/` with the same names, signatures and JSON output (constitution → *Template Exceptions* no. 4). Every
> mention of go-common in these rules means that local package. Don't add dependencies on private modules.

## Architecture (Clean Architecture) — [01-architecture](docs/rules/01-architecture.md)
```
handler  ──►  usecase  ──►  repository / service
(HTTP)        (business)     (DB / external API)
   │             │                 │
   └──── all depend only on interfaces in domain/ ────┘
```
1. `domain/` **must not** import `internal/`, gorm, fiber or go-common. It holds: entity, valueobject, errorcode, dto, interfaces (repository, service, usecase).
2. A handler only: parses the request → converts it to `usecase.XxxInput` → calls the usecase → maps to a response DTO. **No business logic, no DB access.**
3. The usecase holds **all** business logic: validation, rules, transactions. It knows nothing about HTTP/Fiber.
4. A repository only accesses data. No business rules, never returns `apperr`.
5. Dependency injection **only** in `internal/bootstrap/setup.go`. Routes **only** in `internal/routes/router.go`.

## File structure — [02-naming](docs/rules/02-naming.md)
- **1 file = 1 endpoint/use case/repository method.** `create_asset_import.go`, `find_by_asset_ids.go`, etc. Every file has a matching `_test.go`.
- Every package has an `init.go` (struct + `NewXxx` constructor) and an optional `helpers.go`.
- Implementation package names: `<module>_usecase`, `<module>_repository`; handlers: `<module>`. Import alias = package name.
- Method names are the same in every layer: `GetListXxx`, `GetXxxDetail`, `CreateXxx`, `UpdateXxx`, `DeleteXxx`.
- JSON: **camelCase**. DB tables/columns: **snake_case**, singular tables. Error codes: `SCREAMING_SNAKE_CASE`.

## Handler — [03-handler](docs/rules/03-handler.md)
- Parse with `ctx.QueryParser` / `ctx.ParamsParser` / `ctx.BodyParser` into a struct in `domain/dto`. A parse failure → `apperr.BadRequest(errorcode.InvalidRequest, ...)`.
- Convert types in the handler (UUIDs, enums via `valueobject.ParseXxx`) before they reach the usecase.
- Use `ctx.UserContext()` (not the Fiber `ctx`) when calling the usecase.
- Success: `response.Success(ctx, status, "Successfully <action> <module>", dto)`. Delete → `204`. Create → `201`.
- Errors: just `return err`. **Don't** write error responses by hand.

## Usecase — [04-usecase](docs/rules/04-usecase.md)
- Signature: `func (u xxxUsecase) Method(ctx, input usecase.MethodInput) (usecase.MethodOutput, error)`.
- Required order: **log setup → input validation → business rules → writes (in a transaction when >1 write query / read-then-write) → success log**.
- `input.Validate()` (in the domain) returns a plain `error`; the usecase wraps it as `apperr.BadRequest(errorcode.InvalidRequest, err.Error())`.
- A repository `nil, nil` = not found → the usecase decides on `apperr.NotFound(...)`.
- Transactions: `u.txManager.WithinTx(ctx, func(ctx, trx) error {...})`, passing `trx` to **every** repository call inside it.
- Time from `u.now()` (injected), IDs from `uuid.New()`.

## Repository & DB — [05-repository-database](docs/rules/05-repository-database.md)
- Signature: `(ctx context.Context, trx repository.Tx, ...)`. Get the connection through `gormtx.Conn(ctx, r.db, trx)`.
- GORM models (`internal/repository/psql/model`) **never** leave the repository layer — convert through `mapper.ToEntityXxx / ToModelXxx`.
- Not found → `return nil, nil` (using `errors.Is(err, gorm.ErrRecordNotFound)`).
- **Don't use `db.Save`** for updates (it falls back to an INSERT). Use `Model(...).Select(columns...).Updates(...)`.
- **Don't** name fields `CreatedBy`/`UpdatedBy` unless they hold a UUID — the go-common audit plugin overwrites/drops them. Use `CreatedByName`/`UpdatedByName`.
- Every user value → a `?` placeholder. ORDER BY → `querysafety.SafeOrderClause` + an allow-list. LIKE → `querysafety.ContainsPattern`.
- Schema changes → a new migration file in `migrations/` (`NNNNNN_description.up.sql` + `.down.sql`). Never change a released migration.

## Errors — [06-error-handling](docs/rules/06-error-handling.md)
- Business errors **only** through `internal/apperr` (`BadRequest`, `Forbidden`, `NotFound`, `Conflict`, `UnprocessableEntity`).
- A new code → add it to `domain/errorcode` (+ `All`) **and** one English message to the `apperr` catalog (this project: an English-only catalog, constitution → *Template Exceptions* no. 6). A test fails if you forget.
- Technical errors (DB, network) are returned as they are → automatically a 500 "Internal Server Error" (no details leak to the client; they are still logged).

## Security — [07-security](docs/rules/07-security.md)
- This project has no auth: business routes are public under `/api`, without InternalSecret/Identity/JWT and without permission checks.
  If users are ever introduced, go back to the template rules in [07-security](docs/rules/07-security.md).
- No secrets in code, `.env`, `values.yml`, logs or commits. Config must be validated in `config.Validate()`.
- Don't log passwords/tokens/PII/raw request bodies.

## Testing — [08-testing](docs/rules/08-testing.md)
- Every logic file has a test. Table-driven, mocks **per case** (mockery `EXPECT()`), `StubWithinTx` for transactions.
- Usecases: every branch (validation, not found, conflict, repo error, rollback, success). `internal/usecase` coverage ≥ **80%** (gate on main/staging).
- Repositories: sqlmock (regexp) — success, DB error, trx given, wrong Tx type.
- Handlers: `app.Test` + `utils.NewErrorHandler` — status code, body mapping, parse errors.
- After changing a `domain/` interface: `make mocks`. Don't edit generated mock files.

## Logging — [09-logging-observability](docs/rules/09-logging-observability.md)
- `log := logger.WithCtx(ctx, u.log).WithField("usecase", "MethodName")` at the start of every usecase.
- `Error` + `WithError(err)` for technical failures, `Info("success")` at the end. Don't log errors that are already a 4xx `apperr` as Error.

## Key commands
| Purpose | Command |
|---|---|
| Run locally | `make run` |
| Format + vet + test | `make check` |
| Lint | `make lint` |
| Regenerate mocks after changing a `domain/` interface | `make mocks` |
| Usecase coverage (gate ≥ 80%) | `make cover` |

## Adding a new feature
Follow the step-by-step checklist in **[docs/adding-a-feature.md](docs/adding-a-feature.md)**.
