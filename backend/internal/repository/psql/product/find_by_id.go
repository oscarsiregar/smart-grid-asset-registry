package product_repository

import (
	"context"
	"errors"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/repository/psql/mapper"
	"smart-grid-asset-registry/internal/repository/psql/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FindByID returns (nil, nil) when no row matches — the usecase decides
// whether that is a 404.
func (r productRepository) FindByID(ctx context.Context, trx repository.Tx, id uuid.UUID) (*entity.Product, error) {
	db, err := gormtx.Conn(ctx, r.db, trx)
	if err != nil {
		return nil, err
	}

	var product model.Product
	err = db.Where("id = ?", id).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToEntityProduct(&product), nil
}
