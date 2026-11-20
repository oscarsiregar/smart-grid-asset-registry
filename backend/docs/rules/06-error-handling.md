# 06 — Error Handling

## Flow

```
usecase: return apperr.NotFound(errorcode.ProductNotFound)
   │  (apperror.AppError{Code:404, Message:"Product not found"})
handler: return err
   │
bootstrap/fiber.go → ErrorHandler = utils.NewErrorHandler(log)   (go-common)
   ├─ *apperror.AppError → status = Code, message = Message   (logged when ≥ 500)
   ├─ *fiber.Error       → status = Code
   └─ any other error    → 500 "Internal Server Error" (details are NOT sent, but they are logged)
```

As a result: technical errors (a dropped DB connection, a wrong query) **never leak** to the
client, and business errors always have the right status & message.

## When to use what

| Situation | Function | HTTP |
|---|---|---|
| Body/query/path cannot be parsed, wrong UUID/enum, `Validate()` fails | `apperr.BadRequest(errorcode.InvalidRequest, detail)` | 400 |
| No permission / not the owner of the data | `apperr.Forbidden(errorcode.PermissionDenied)` or a specific code | 403 |
| The referenced data does not exist | `apperr.NotFound(errorcode.XxxNotFound)` | 404 |
| State conflict: a duplicate, a changed version | `apperr.Conflict(errorcode.XxxAlreadyExists, arg)` | 409 |
| Valid input that breaks a business rule (wrong status, no longer applicable) | `apperr.UnprocessableEntity(errorcode.Xxx, arg)` | 422 |
| A technical failure | `return err` as it is | 500 |

401 (token) and the BFF 403 (secret) are produced by the go-common middleware — usecases do not need to handle them.

## Adding a new error code

1. `domain/errorcode/errorcode.go` — add the constant **and** put it in the `All` slice.
2. `internal/apperr/apperr.go` — add a catalog entry:

   ```go
   errorcode.ProductCodeAlreadyExists: "Product code %s is already in use",
   ```

   (In the template each entry holds an Indonesian (`ID`) and an English (`EN`) message.)
3. `TestCatalog_EveryErrorCodeHasAMessage` makes sure step 2 is not missed.

> **This project:** the catalog holds English messages only (constitution → *Template Exceptions* no. 6), as in
> the example above.

## Message rules

- Messages to the client are polite and something the user can act on ("Product code P-001 is already in use"), not technical jargon. The template writes them in Indonesian; this project in English.
- `%s` arguments only for values that are safe to show (codes, ids the user sent). Never insert DB error text, queries, stack traces or another user's data.
- One code = one meaning. A specific code (`PERIOD_NOT_EDITABLE`) is better than a generic code with varying messages.
- Usecases & handlers **must not** call go-common's `apperror.*` directly — always go through `apperr`, so messages stay in one place.

## Errors from the domain

`Input.Validate()` and `valueobject.ParseXxx` return a plain `error` (the domain does not
know `apperr`). The calling layer wraps it:

```go
if err := input.Validate(); err != nil {
	return Output{}, apperr.BadRequest(errorcode.InvalidRequest, err.Error())
}
```

Because the message from `Validate()` is sent to the client, keep it short and safe
(`"code is required"`, `"price must not be negative"`).

## Panics

- `middleware.RecoverConfig()` (installed first in `main.go`) turns a panic into a 500 for that request only — the service stays up.
- A panic is **not** a control-flow mechanism. Don't `panic` for an error that can be returned.
- `log.Fatal` is only allowed in `main.go` / bootstrap during startup.

## Wrapping

When adding context to a technical error use `%w`, so `errors.Is/As` keep working:

```go
return nil, fmt.Errorf("user-management: get user %s: %w", id, err)
```

Don't compare errors with `==`; use `errors.Is` (the `errorlint` linter is on).
