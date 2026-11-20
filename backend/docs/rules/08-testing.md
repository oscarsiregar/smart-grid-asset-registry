# 08 — Testing

## Targets

| Layer | Tool | What must be tested | Coverage |
|---|---|---|---|
| `domain/valueobject` | testify | Parse (valid/invalid), Scan (string, []byte, wrong type, wrong value), Value, String, IsValid | 100% |
| `domain/usecase` (Validate/SetDefault) | testify, table-driven | every validation rule & default | 100% |
| `internal/usecase` | mockery mocks + `StubWithinTx` | **every branch**: success, permission denied, validation, not found, conflict, every repo error, rollback | **≥ 80% (CI gate)**, target 100% |
| `internal/repository/psql` | sqlmock (regexp) | the generated SQL, success, DB error, trx given, wrong Tx type, not found → nil,nil | high |
| `internal/handler` | `fiber.App.Test` + usecase mock | parse error → 400, input passed on correctly, response mapping, success status, AppError → status, plain error → 500 | high |
| mapper / helpers / querysafety | testify | every field, nil, empty slice | 100% |

Not tested (excluded in CI): `cmd/`, `internal/bootstrap`, `internal/routes`, `internal/mocks`, declaration-only packages (`domain/repository|service|dto|entity|errorcode`, `psql/model`).

> **This project:** repositories and usecases are also tested against a real PostgreSQL in
> `internal/integrationtest/` and `internal/testsupport/pgtest/`, skipped without `TEST_DATABASE_URL`
> (constitution → *Template Exceptions* no. 11).

## Usecase test pattern (table-driven, mocks per case)

```go
tests := []struct {
	name    string
	input   usecase.CreateProductInput
	setup   func(d deps)          // set the EXPECT() calls for this case
	wantErr error
}{
	{
		name:  "error code already exists rolls back",
		input: validInput,
		setup: func(d deps) {
			d.withinTx()                                    // WithinTx runs the callback with d.trx
			d.repo.EXPECT().FindByCode(mock.Anything, d.trx, "P-001").Return(sampleProduct(), nil)
			d.trx.EXPECT().Rollback().Return(nil)           // makes sure the rollback happens
		},
		wantErr: apperr.Conflict(errorcode.ProductCodeAlreadyExists, "P-001"),
	},
}
for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		d := newDeps(t)       // NEW mocks for every case
		tt.setup(d)
		output, err := d.usecase().CreateProduct(ctx, tt.input)
		assert.Equal(t, tt.wantErr, err)
	})
}
```

Rules:
1. **Mocks are created per case** (`newDeps(t)`). Mockery registers `AssertExpectations` through `t.Cleanup`, so an expected call that does not happen → fails; an unexpected call → panics/fails. This is what proves "the repository is not called when permission is denied".
2. Use `EXPECT()` (typed), not `On("Method")`.
3. Check the transaction: the `d.trx` argument (not `mock.Anything`) for repos that must run inside the transaction, and `Commit()` / `Rollback()` expected explicitly.
4. Pin the time: `now: func() time.Time { return fixedNow }`.
5. Match the entity sent to the repo with `mock.MatchedBy(func(p *entity.Product) bool {...})` to verify the fields the usecase computes (trim, audit, timestamps).
6. Compare errors with `assert.Equal(t, tt.wantErr, err)` — `apperr` produces comparable values, so the code & message are verified too.

## Repository test pattern (sqlmock)

```go
gdb, mock := newMockDB(t) // regexp matcher; ExpectationsWereMet is checked in Cleanup
mock.ExpectQuery(`SELECT \* FROM "product" WHERE id = \$1 ORDER BY "product"\."id" LIMIT \$2`).
	WithArgs(id, 1).
	WillReturnRows(productRow(sqlmock.NewRows(productColumns), id, "P-001"))
```

- Write the expected SQL **as completely as possible** (not just `SELECT`) — this is what catches regressions such as `Save` → upsert or zero values that are not updated.
- GORM wraps writes in a transaction by default: `ExpectBegin` → `ExpectExec` → `ExpectCommit` / `ExpectRollback`.
- Required cases per method: success, DB error, `wrapTx(gdb)`, `unsupportedTx{}`; plus `nil, nil` for not found.

## Handler test pattern

```go
app := newTestApp(NewProductHandler(uc), identityWith(valueobject.PermissionCreateProduct))
resp := doRequest(t, app, "POST", "/products", body)
require.Equal(t, fiber.StatusCreated, resp.StatusCode)
```

- `newTestApp` uses the same `utils.NewErrorHandler` as production, and a middleware that injects the identity like `Identity()`.
- For invalid input, create a usecase mock **without** EXPECT — this proves the usecase is not called.
- Verify the JSON shape (camelCase, `items` not `null`, RFC3339 time format).

## Mocks

- Generated from the `domain/` interfaces by mockery (`.mockery.yaml`): `make mocks`.
- **Don't edit** generated files. Hand-written helpers (e.g. `txhelpers.go`) carry a "Hand-written" comment and are not overwritten by mockery.
- Commit the generated mocks together with the interface change.

## Commands

```sh
make test     # unit tests (the same packages as CI)
make cover    # internal/usecase coverage, 80% gate
go test ./internal/usecase/product/ -run TestCreateProduct -v   # a single test
```

## End-to-end smoke test (before a major release)

1. Start a temporary Postgres and apply `migrations/*.up.sql`.
2. `make run` with a `.env` that points at that DB.
3. Call the endpoints with the `X-Internal-Secret` header and `Authorization: Bearer <HS256 JWT with id, name, permissions>`.
4. Verify: 403 without the secret, 401 without a JWT, 403 without the permission, the CRUD flow, 409 on a duplicate, 404 after a delete.
