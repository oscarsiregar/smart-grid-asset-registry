# 04 — Usecase (Business Logic)

A usecase is the **only** place where business decisions are made: who may do what,
which data is valid, what happens on a conflict, and what must be atomic.

## Contract (in `domain/usecase/<module>.go`)

```go
type ProductUsecase interface {
	CreateProduct(ctx context.Context, input CreateProductInput) (CreateProductOutput, error)
	...
}

type CreateProductInput struct {
	Code   string
	...
	ActorName   string   // for the *_by_name columns
	Permissions []string // from the caller's identity
}

type CreateProductOutput struct {
	Product *entity.Product
}

// Validate: the shape of the input only (required, length, range, valid enum).
// Returns a plain error — the domain does not know apperr.
func (i CreateProductInput) Validate() error { ... }

// SetDefault: default paging/sorting, the maximum limit.
func (i *GetListProductInput) SetDefault() { ... }
```

- Input/Output are always **value** structs named `<Method>Input` / `<Method>Output` — even with a single field. Adding a field later does not change the signature.
- A method without returned data (delete) may return just `error`.
- The Output carries **entities**, not DTOs. Mapping to JSON is the handler's job.

## Order of steps (MUST)

```go
func (pu productUsecase) UpdateProduct(ctx context.Context, input usecase.UpdateProductInput) (usecase.UpdateProductOutput, error) {
	// 1. a logger with context
	log := logger.WithCtx(ctx, pu.log).WithField("usecase", "UpdateProduct")

	// 2. authorisation — FIRST, before any validation or query
	if !hasPermission(input.Permissions, valueobject.PermissionUpdateProduct) {
		log.Warn("permission denied")
		return usecase.UpdateProductOutput{}, apperr.Forbidden(errorcode.PermissionDenied)
	}

	// 3. validate the shape of the input
	if err := input.Validate(); err != nil {
		return usecase.UpdateProductOutput{}, apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	// 4-5. business rules + writes, atomically
	err := pu.txManager.WithinTx(ctx, func(ctx context.Context, trx repository.Tx) error {
		product, err := pu.productRepository.FindByID(ctx, trx, input.ID)
		if err != nil { log.WithError(err).Error("failed to find product by id"); return err }
		if product == nil { return apperr.NotFound(errorcode.ProductNotFound) }
		... // check conflicts, mutate the entity
		return pu.productRepository.Update(ctx, trx, product)
	})
	if err != nil {
		return usecase.UpdateProductOutput{}, err
	}

	// 6. a success log with the relevant id
	log.WithField("productID", input.ID).Info("success")
	return usecase.UpdateProductOutput{Product: product}, nil
}
```

## Authorisation

- One permission per endpoint, defined in `domain/valueobject/permission.go`.
- Default **deny**: no permission = `403`.
- Data-based authorisation (e.g. only Sharia-category debtors, only the assigned reviewer)
  is also in the usecase, **after** the data is read: use paired permissions
  (`xxx_sharia` / `xxx_conventional`) + the `hasScopedPermissions` / `isContractCategoryPermitted` helpers
  as in the reference repo, and a specific `Forbidden` error.
- Don't rely on the frontend / BFF hiding buttons as authorisation.

## Validation: two levels

| Kind | Where | Error |
|---|---|---|
| Shape (required, length, range, enum) | `Input.Validate()` in the domain | `400 INVALID_REQUEST` |
| Rules that need data (unique, matching status, existing relation) | usecase | `404` / `409` / `422` with a specific errorcode |

- Normalisation (trim, upper-casing codes) happens in the usecase before saving, and is used consistently for the uniqueness check.
- A uniqueness check in the usecase gives a friendly 409; **a unique constraint in the DB is still required** as the guarantee under concurrent requests.

## Transactions

Use `WithinTx` when:
- there is **more than one** write query, or
- there is a **read-then-write** pattern whose result depends on the read (check unique then insert, check status then update).

Rules:
1. Every repository call inside the callback **must** receive `trx` (not `nil`).
2. Return an error from the callback → automatic rollback. Don't call `Commit/Rollback` by hand.
3. Don't call external APIs inside a transaction (it holds a DB connection). Fetch external data **before** `WithinTx`.
4. Results that need to be returned are captured in variables outside the closure (`var product *entity.Product`).
5. A single write query does not need a transaction (see `DeleteProduct`).

## Other

- Time: `pu.now()` (a `now func() time.Time` field set to `time.Now` in the constructor) — so tests can pin the time.
- New IDs: `uuid.New()` in the usecase, not a DB default, so the entity is complete before it is saved and can be returned.
- Paging: `input.SetDefault()`, then pass it on to `repository.FindXxxFilter`. The maximum limit is `usecase.MaxLimit` (100).
- A usecase does not call another usecase. Shared code → `helpers.go` in that package.
- Private functions called by several usecases are still tested in `helpers_test.go`.
