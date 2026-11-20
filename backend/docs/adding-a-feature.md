# Adding a Module / Endpoint — Checklist

Example: creating a **`supplier`** module with CRUD. Work in order from the inside
(domain) out (routes); every step has a counterpart in the `product` module — copy it
and adapt.

## 0. Schema

- [ ] `migrations/00000N_create_supplier.up.sql` + `.down.sql` (singular table, UUID PK, `*_by_name` audit columns, indexes for filter/sort/unique columns).

## 1. Domain (`domain/`)

- [ ] `valueobject/supplier_status.go` (+ `_test.go`) if there is an enum — copy `product_status.go`.
- [ ] `valueobject/permission.go` — add `PermissionGetListSupplier = "supplier.get_list"`, etc. (one per endpoint).
- [ ] `errorcode/errorcode.go` — `SupplierNotFound`, … **and** add them to `All`.
- [ ] `entity/supplier.go` — a plain struct.
- [ ] `repository/supplier.go` — interface + `FindSupplierFilter`.
- [ ] `usecase/supplier.go` — interface + `XxxInput/XxxOutput` + `Validate()` / `SetDefault()` (+ `supplier_test.go`).
- [ ] `dto/supplier.go` — `XxxRequest` + `SupplierResponse`.

## 2. Error messages

- [ ] `internal/apperr/apperr.go` — an ID + EN catalog entry for every new code.

## 3. Repository (`internal/repository/psql/`)

- [ ] `model/supplier.go` (gorm tags + `TableName()`).
- [ ] `mapper/supplier.go` (+ `_test.go`).
- [ ] `supplier/init.go` (`NewSupplierRepository`, `sortableColumns`), then one file per method + `_test.go` (sqlmock), `testdb_test.go`.

## 4. Mocks

- [ ] `make mocks` → `internal/mocks/repository/SupplierRepository.go`, `internal/mocks/usecase/SupplierUsecase.go`.

## 5. Usecase (`internal/usecase/supplier/`)

- [ ] `init.go` (struct + `NewSupplierUsecase`, `now` field), `helpers.go`.
- [ ] One file per use case, in this order: log → permission → validate → rules → write (tx) → success log.
- [ ] A `_test.go` per file + `helpers_test.go` (`deps`, `newDeps`, `withinTx`, fixtures).

## 6. Handler (`internal/handler/supplier/`)

- [ ] `init.go` (`SupplierHandler` interface + constructor), `helpers.go` (`caller`, `toSupplierResponse`).
- [ ] One file per endpoint + `_test.go`, `testutil_test.go`.

## 7. Wiring

- [ ] `internal/bootstrap/setup.go` — repository → usecase → handler, added to `RouteDependencies`.
- [ ] `internal/routes/router.go` — a field in `RouteDependencies` + the `authenticated.Group("/suppliers")` group.

## 8. Documentation & verification

- [ ] `docs/api/supplier.md` — endpoints, permissions, request/response, errors.
- [ ] Register the new permissions with the identity service / role mapping.
- [ ] `make check` and `make lint` clean; usecase coverage ≥ 80%.
- [ ] (Optional) an end-to-end smoke test — see [rules/08-testing.md](rules/08-testing.md).

## Adding an endpoint to an existing module

Step 1 (permission, errorcode, Input/Output, DTO) → 2 → a repository method if needed
(3) → `make mocks` → usecase + test (5) → handler + test (6) → the method on the handler
interface & the route (7) → docs (8).

## Adding an external integration

1. `domain/service/<name>.go` — interface + Input/Output.
2. `.mockery.yaml` — enable the `smart-grid-asset-registry/domain/service` package (dir `internal/mocks/service`).
3. `internal/service/<name>/init.go` (the constructor takes the base URL, secret and logger; an HTTP client with a timeout) + one file per call + tests with `httptest.NewServer`.
4. New env variables in `config.go` (+ `Validate`), `.env.example`, `values.yml`/Secret.
5. Inject it in `bootstrap/setup.go` into the usecases that need it.
