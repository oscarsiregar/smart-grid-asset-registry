package product_repository

import (
	"context"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/repository/psql/mapper"
)

func (r productRepository) Create(ctx context.Context, trx repository.Tx, product *entity.Product) error {
	db, err := gormtx.Conn(ctx, r.db, trx)
	if err != nil {
		return err
	}

	return db.Create(mapper.ToModelProduct(product)).Error
}
