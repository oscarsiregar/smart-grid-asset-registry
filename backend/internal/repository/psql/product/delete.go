package product_repository

import (
	"context"

	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/repository/psql/model"

	"github.com/google/uuid"
)

func (r productRepository) Delete(ctx context.Context, trx repository.Tx, id uuid.UUID) error {
	db, err := gormtx.Conn(ctx, r.db, trx)
	if err != nil {
		return err
	}

	return db.Where("id = ?", id).Delete(&model.Product{}).Error
}
