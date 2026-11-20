# 02 — Naming & File Structure

## One file = one unit

| Layer | One file holds | Example |
|---|---|---|
| handler | 1 endpoint | `internal/handler/product/create_product.go` |
| usecase | 1 use case (interface method) | `internal/usecase/product/create_product.go` |
| repository | 1 repository method | `internal/repository/psql/product/find_by_code.go` |
| service | 1 external API call | `internal/service/user_management/get_user_by_id.go` |
| valueobject | 1 enum | `domain/valueobject/product_status.go` |
| other domain | 1 file per module | `domain/dto/product.go`, `domain/usecase/product.go` |

Every file above that contains logic **must** have a matching `_test.go` with the same name.

Supporting files per package:

- `init.go` — the handler interface (handlers only), the implementation struct, the `NewXxx` constructor.
- `helpers.go` — private functions shared within the package (`hasPermission`, `toProductResponse`).
- `testutil_test.go` / `testdb_test.go` / `helpers_test.go` — test-only helpers.

## Packages & import aliases

| Location | Package name | Import alias in bootstrap |
|---|---|---|
| `internal/handler/product` | `product` | (no alias) |
| `internal/usecase/product` | `product_usecase` | `product_usecase` |
| `internal/repository/psql/product` | `product_repository` | `product_repository` |
| `internal/service/user_management` | `user_management_service` | `user_management_service` |
| mock | `mocks` | `mocksrepo`, `mocksusecase`, `mockssvc` |
| go-common fiber middleware | — | `pkg_middleware_fiber` |

Module folder name = the aggregate name in singular snake_case (`product`, `covenant_config`).

## Method names (the same in ALL layers)

| Operation | Usecase / Handler | Repository |
|---|---|---|
| List + paging | `GetListProduct` | `Find(filter)` → `([]*E, total, error)` |
| Detail | `GetProductDetail` | `FindByID`, `FindByCode`, `FindByXxx` |
| All without paging | — | `FindAll` |
| Create | `CreateProduct` | `Create` |
| Update | `UpdateProduct` | `Update` |
| Delete | `DeleteProduct` | `Delete`, `DeleteByXxx` |
| Domain action | verb + object: `ApprovePeriod`, `SubmitPeriodForVerification` | — |

The names of the handler, usecase, Input/Output, DTO, permission and file must **line up**:

```
CreateProduct  →  create_product.go
               →  usecase.CreateProductInput / CreateProductOutput
               →  dto.CreateProductRequest  (common response: dto.ProductResponse)
               →  valueobject.PermissionCreateProduct = "product.create"
```

## Types

- Implementation structs are **private** (`productUsecase`); public constructors return the **interface** (`usecase.ProductUsecase`).
- Receiver: an abbreviation of the struct name (`pu productUsecase`, `r productRepository`, `h *productHandler`). Usecase & repository receivers by value; handlers by pointer.
- Optional/nullable fields → pointers (`*string`, `*int`, `*valueobject.X`). Required fields → plain values.
- IDs: `uuid.UUID` in entities/usecases; `string` in DTOs.
- Money/numbers: `float64` (DB `NUMERIC`). When decimal precision is critical (accounting), agree on a decimal type at project level.
- Time: `time.Time` in entities; an RFC3339 string in response DTOs.

## String conventions

| Item | Format | Example |
|---|---|---|
| JSON field | camelCase | `totalItems`, `createdByName` |
| Query param | camelCase | `?search=&page=&limit=&sort=&order=` |
| Path | kebab-case, plural | `/v1/products/:id`, `/v1/report-types` |
| DB table | snake_case **singular** | `product`, `debtor_covenant_config` |
| DB column | snake_case | `created_at`, `updated_by_name` |
| Index | `ix_<table>_<column>` / unique `ux_<table>_<column>` | `ux_product_code` |
| Enum value | SCREAMING_SNAKE_CASE | `ACTIVE`, `PENDING_APPROVAL` |
| Error code | SCREAMING_SNAKE_CASE | `PRODUCT_NOT_FOUND` |
| Permission | `<module>.<action>` snake_case | `product.get_list`, `period.approve_sharia` |
| Env var | SCREAMING_SNAKE_CASE | `DB_MAX_OPEN_CONNS` |
| Success message | `"Successfully <action> <module>"` | `"Successfully get list product"` |

## Comments & language

- Code, identifiers, code comments: **English**.
- Error messages to the client: **Indonesian** (the `apperr` catalog, with EN equivalents).
- Documentation (`docs/`, README): Indonesian.
- Comments explain **why**, not what. Comments about incidents/fixes include the date & the reason.

> **This project:** error messages are English only (constitution → *Template Exceptions* no. 6), and the
> documentation is English (no. 9).
