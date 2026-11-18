package product_repository

import (
	"context"

	"backend-template/domain/entity"
	"backend-template/domain/repository"
	"backend-template/internal/repository/psql/gormtx"
	"backend-template/internal/repository/psql/mapper"
)

func (r productRepository) Create(ctx context.Context, trx repository.Tx, product *entity.Product) error {
	db, err := gormtx.Conn(ctx, r.db, trx)
	if err != nil {
		return err
	}

	return db.Create(mapper.ToModelProduct(product)).Error
}
