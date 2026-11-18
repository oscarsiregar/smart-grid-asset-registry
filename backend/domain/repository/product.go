package repository

import (
	"context"

	"backend-template/domain/entity"
	"backend-template/domain/valueobject"

	"github.com/google/uuid"
)

// ProductRepository: every method takes ctx and an optional trx (nil = run
// outside a transaction). "Not found" is (nil, nil), never an error.
type ProductRepository interface {
	Create(ctx context.Context, trx Tx, product *entity.Product) error
	Update(ctx context.Context, trx Tx, product *entity.Product) error
	Delete(ctx context.Context, trx Tx, id uuid.UUID) error
	FindByID(ctx context.Context, trx Tx, id uuid.UUID) (*entity.Product, error)
	FindByCode(ctx context.Context, trx Tx, code string) (*entity.Product, error)
	// Find returns one page of products plus the total matching row count.
	Find(ctx context.Context, trx Tx, filter FindProductFilter) ([]*entity.Product, int64, error)
}

type FindProductFilter struct {
	Search string                     // matched against code and name (ILIKE)
	Status *valueobject.ProductStatus // nil = any status
	Page   int                        // 1-based
	Limit  int                        // < 0 = no limit
	Sort   string                     // column name, sanitized by querysafety
	Order  string                     // ASC | DESC
}
