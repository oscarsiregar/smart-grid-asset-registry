package product_repository

import (
	"context"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/repository/psql/mapper"
)

// Update writes every mutable column (zero values included) of the row
// identified by product.ID. created_at/created_by_name are never touched.
//
// Do NOT use db.Save here: when no row matches, GORM's Save silently falls
// back to an INSERT (upsert).
func (r productRepository) Update(ctx context.Context, trx repository.Tx, product *entity.Product) error {
	db, err := gormtx.Conn(ctx, r.db, trx)
	if err != nil {
		return err
	}

	return db.Model(mapper.ToModelProduct(product)).
		Select("code", "name", "description", "price", "status", "updated_at", "updated_by_name").
		Updates(mapper.ToModelProduct(product)).Error
}
