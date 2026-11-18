package product_repository

import (
	"context"

	"backend-template/domain/repository"
	"backend-template/internal/repository/psql/gormtx"
	"backend-template/internal/repository/psql/model"

	"github.com/google/uuid"
)

func (r productRepository) Delete(ctx context.Context, trx repository.Tx, id uuid.UUID) error {
	db, err := gormtx.Conn(ctx, r.db, trx)
	if err != nil {
		return err
	}

	return db.Where("id = ?", id).Delete(&model.Product{}).Error
}
