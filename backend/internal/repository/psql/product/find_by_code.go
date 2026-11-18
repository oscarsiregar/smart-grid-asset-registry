package product_repository

import (
	"context"
	"errors"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/repository/psql/mapper"
	"smart-grid-asset-registry/internal/repository/psql/model"

	"gorm.io/gorm"
)

func (r productRepository) FindByCode(ctx context.Context, trx repository.Tx, code string) (*entity.Product, error) {
	db, err := gormtx.Conn(ctx, r.db, trx)
	if err != nil {
		return nil, err
	}

	var product model.Product
	err = db.Where("code = ?", code).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToEntityProduct(&product), nil
}
