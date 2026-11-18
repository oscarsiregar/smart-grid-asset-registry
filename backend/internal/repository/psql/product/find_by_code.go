package product_repository

import (
	"context"
	"errors"

	"backend-template/domain/entity"
	"backend-template/domain/repository"
	"backend-template/internal/repository/psql/gormtx"
	"backend-template/internal/repository/psql/mapper"
	"backend-template/internal/repository/psql/model"

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
