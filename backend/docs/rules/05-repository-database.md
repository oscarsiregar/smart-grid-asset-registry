# 05 — Repository & Database

## Interface (in `domain/repository/<module>.go`)

```go
type ProductRepository interface {
	Create(ctx context.Context, trx Tx, product *entity.Product) error
	Update(ctx context.Context, trx Tx, product *entity.Product) error
	Delete(ctx context.Context, trx Tx, id uuid.UUID) error
	FindByID(ctx context.Context, trx Tx, id uuid.UUID) (*entity.Product, error)
	FindByCode(ctx context.Context, trx Tx, code string) (*entity.Product, error)
	Find(ctx context.Context, trx Tx, filter FindProductFilter) ([]*entity.Product, int64, error)
}
```

- The first parameter is always `ctx`, the second always `trx Tx` (may be `nil`).
- Takes & returns **entities**, never GORM models.
- More than 3 filter parameters → a `FindXxxFilter` struct in the `repository` package.

## Implementation

```go
func (r productRepository) FindByID(ctx context.Context, trx repository.Tx, id uuid.UUID) (*entity.Product, error) {
	db, err := gormtx.Conn(ctx, r.db, trx) // trx if given, otherwise r.db.WithContext(ctx)
	if err != nil {
		return nil, err
	}

	var product model.Product
	err = db.Where("id = ?", id).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // not found is NOT an error
		}
		return nil, err
	}
	return mapper.ToEntityProduct(&product), nil
}
```

### MUST

1. The connection always comes from `gormtx.Conn(ctx, r.db, trx)` — this guarantees `ctx` (timeout, trace, audit user id) goes along and the transaction is used when given.
2. **Not found = `nil, nil`** for a single `FindByXxx`; an empty list = an empty, non-nil slice.
3. DB errors are returned as they are (not wrapped in `apperr`, not logged — the usecase logs).
4. GORM models live only in `psql/model`, converted through `psql/mapper` (`ToModelXxx`, `ToEntityXxx`, `ToEntityXxxs` → non-nil slice).
5. A repository does not call another repository and does not open its own transaction.

### MUST NOT

| Don't | Why | Use |
|---|---|---|
| `db.Save(m)` for an update | If the row does not exist, GORM silently INSERTs (upsert) | `db.Model(m).Select("column", ...).Updates(m)` — explicit, zero values are saved too |
| `db.Updates(struct)` without `Select` | zero values (`0`, `""`, `false`, `nil`) are **skipped** | add `Select(...)` |
| `fmt.Sprintf` / concatenating user values into SQL | SQL injection | `?` placeholders |
| raw `Order(request.Sort)` | ORDER BY cannot be parameterised | `querysafety.SafeOrderClause(sort, order, default, defaultOrder, allowList...)` |
| `"%"+q+"%"` for LIKE | a user's `%`/`_` becomes a wildcard | `querysafety.ContainsPattern(q)` |
| A model field named `CreatedBy`/`UpdatedBy` holding a name | the go-common audit plugin overwrites it with the JWT user UUID and **drops** non-UUID values | `CreatedByName`/`UpdatedByName` (or a UUID-typed `CreatedBy` when automatic auditing is wanted) |
| `db.Use(database.NewAuditPlugin())` | already registered by `database.Connect` → `gorm.ErrRegistered` | nothing needed |
| Reusing a GORM chain for COUNT then SELECT | the statements leak into each other | `query = query.Session(&gorm.Session{})` after the WHERE is built |

## List paging

```go
query := db.Model(&model.Product{})
// ... dynamic WHERE
query = query.Session(&gorm.Session{})

var total int64
query.Count(&total)

query.Order(querysafety.SafeOrderClause(filter.Sort, filter.Order, "created_at", "DESC", sortableColumns...)).
	Limit(filter.Limit).Offset((filter.Page - 1) * filter.Limit).
	Find(&rows)
```

- `sortableColumns` = the allow-list of sortable columns, defined in the repository's `init.go`.
- `Limit <= 0` means no limit (internal calls only; usecases always set a default).

## Schema & migrations

- **Singular** snake_case tables; PK `id UUID`; audit columns `created_at`, `created_by_name`, `updated_at`, `updated_by_name` (+ `deleted_at`, `deleted_by_name` for soft delete).
- Enums are stored as `VARCHAR` + `CHECK (status IN (...))`, mapped to a `valueobject` that implements `sql.Scanner` & `driver.Valuer`.
- Money `NUMERIC(18,2)`, time `TIMESTAMPTZ`.
- Every column used for filtering/sorting/uniqueness → an index (`ix_…`, `ux_…`).
- Migrations: `migrations/NNNNNN_description.up.sql` + `.down.sql`, numbered in order, **idempotent** (`IF NOT EXISTS`). Compatible with golang-migrate.
- A migration that is already in `main` **must not be changed** — create a new migration.
- Breaking changes (renaming/dropping a column) are done in steps: add the new column → deploy code that writes both → migrate the data → drop the old column in the next release.

> **This project:** no CHECK constraints — enum and range columns are plain columns validated by the backend
> (constitution → *Template Exceptions* no. 8).

## Soft delete vs hard delete

- The `product` example defaults to **hard delete**.
- When data has audit/legal value (documents, approvals): soft delete — a `status = DELETED` or `deleted_at` column, and **every** `Find*` query must filter out deleted rows. Document the choice in the usecase.

## External integrations (`internal/service`)

- The interface in `domain/service`, the implementation in `internal/service/<name>/` (1 file = 1 call); the constructor takes the base URL & secret from config.
- An HTTP client with an explicit **timeout** (go-common `utils.NewTracingClient(timeout)` forwards the trace id).
- Requests always use `http.NewRequestWithContext(ctx, ...)`; close `resp.Body`.
- Non-2xx → a clear error (`fmt.Errorf("user-management: GET /users/%s: status %d", ...)`); an upstream 404 that means "does not exist" → `nil, nil`.
- No `InsecureSkipVerify: true` in production. If it is unavoidable in dev (self-signed certificates), make it an env setting, default `false`.
- Test with `httptest.NewServer`.
