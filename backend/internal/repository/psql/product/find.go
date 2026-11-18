package product_repository

import (
	"context"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/repository/psql/mapper"
	"smart-grid-asset-registry/internal/repository/psql/model"
	"smart-grid-asset-registry/internal/repository/psql/querysafety"

	"gorm.io/gorm"
)

func (r productRepository) Find(ctx context.Context, trx repository.Tx, filter repository.FindProductFilter) ([]*entity.Product, int64, error) {
	db, err := gormtx.Conn(ctx, r.db, trx)
	if err != nil {
		return nil, 0, err
	}

	query := db.Model(&model.Product{})
	if filter.Search != "" {
		pattern := querysafety.ContainsPattern(filter.Search)
		query = query.Where("(code ILIKE ? OR name ILIKE ?)", pattern, pattern)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	// Session makes the built WHERE reusable for both COUNT and SELECT
	// without one statement leaking into the other.
	query = query.Session(&gorm.Session{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageQuery := query.Order(querysafety.SafeOrderClause(filter.Sort, filter.Order, "created_at", "DESC", sortableColumns...))
	if filter.Limit > 0 {
		page := filter.Page
		if page < 1 {
			page = 1
		}
		pageQuery = pageQuery.Limit(filter.Limit).Offset((page - 1) * filter.Limit)
	}

	var products []*model.Product
	if err := pageQuery.Find(&products).Error; err != nil {
		return nil, 0, err
	}
	return mapper.ToEntityProducts(products), total, nil
}
