# 03 — Handler (HTTP)

A handler is a **thin adapter** between Fiber and a usecase. If your handler is longer
than about 50 lines or holds an `if` that is not about parsing, that logic belongs in the usecase.

## Standard anatomy

```go
// CreateProduct godoc
// POST /v1/products
func (h *productHandler) CreateProduct(ctx *fiber.Ctx) error {
	// 1. parse into the request DTO
	var request dto.CreateProductRequest
	if err := ctx.BodyParser(&request); err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	// 2. convert types (UUID, enum) — a failure is a 400
	status, err := valueobject.ParseProductStatus(request.Status)
	if err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	// 3. the caller's identity
	actorName, permissions := caller(ctx)

	// 4. call ONE usecase with ctx.UserContext()
	output, err := h.productUsecase.CreateProduct(ctx.UserContext(), usecase.CreateProductInput{...})
	if err != nil {
		return err // the global ErrorHandler formats it
	}

	// 5. map Output → response DTO
	return response.Success(ctx, fiber.StatusCreated, "Successfully create product", toProductResponse(output.Product))
}
```

## Rules

1. **Parsing**: `QueryParser` (GET list), `ParamsParser` (path `:id`), `BodyParser` (POST/PUT/PATCH).
   For a PUT/PATCH with a path param + a body, call `ParamsParser` and then `BodyParser` on the same struct.
2. **Conversion**: UUIDs with `uuid.Parse`, enums with `valueobject.ParseXxx`. The UUID parse error message: `"id must be a valid UUID"` (don't leak parser details).
3. **Identity**: only from `pkg_middleware_fiber.GetUserIdentity(ctx)` (filled in by middleware). A caller without an identity → `Permissions: []string{}`, so the usecase refuses.
   Never read the `X-User-*` headers yourself.
4. **Context**: always `ctx.UserContext()` — it carries the trace id and the user id for the audit plugin. Never pass `*fiber.Ctx` to a usecase.
5. **Success status**:

   | Operation | Status | Body |
   |---|---|---|
   | GET list / detail | 200 | `response.Success` |
   | POST create | 201 | the created entity |
   | PUT / PATCH | 200 | the entity after the change |
   | DELETE | 204 | empty (`ctx.SendStatus`) |
   | Async action | 202 | the job status |

6. **Errors**: `return err`. Don't use `c.Status(...).JSON(...)` for errors. `utils.NewErrorHandler` (in `bootstrap/fiber.go`) maps an `apperror.AppError` to its status and hides any other error as a 500.
7. **List response**: `{ "items": [...], "totalItems": N }`. `items` **must not be `null`** — initialise with `make([]T, 0, n)`.
8. **Mapping** entity → DTO in a `toXxxResponse` helper (one response shape per entity for every endpoint).
9. A handler does **not** call repositories, services or other usecases, and does not call more than one usecase method.

## Response envelope (from go-common)

```json
{ "code": 200, "status": "success", "message": "Successfully get list product", "data": { ... } }
{ "code": 404, "status": "error",   "message": "Product not found",             "data": null }
```

## Route registration

Only in `internal/routes/router.go`, under the `authenticated` group:

```go
productGroup := authenticated.Group("/products")
productGroup.Get("/", deps.ProductHandler.GetListProduct)
productGroup.Post("/", deps.ProductHandler.CreateProduct)
productGroup.Get("/:id", deps.ProductHandler.GetProductDetail)
productGroup.Put("/:id", deps.ProductHandler.UpdateProduct)
productGroup.Delete("/:id", deps.ProductHandler.DeleteProduct)
```

Static routes (`/months`, `/bulk-generate`) are registered **before** parameterised routes (`/:id`) so they are not caught as an id.

## REST conventions

| Need | Method & path |
|---|---|
| list | `GET /v1/<module-plural>` |
| detail | `GET /v1/<module-plural>/:id` |
| create | `POST /v1/<module-plural>` |
| full replace | `PUT /v1/<module-plural>/:id` |
| partial change / action | `PATCH /v1/<module-plural>/:id[/<action>]` (e.g. `/submit`, `/approve`) |
| delete | `DELETE /v1/<module-plural>/:id` |
| sub-resource | `/v1/periods/:id/report-compliance/:catalogId` |
