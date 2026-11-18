package mapper

import (
	"backend-template/domain/entity"
	"backend-template/internal/repository/psql/model"
)

func ToModelProduct(e *entity.Product) *model.Product {
	if e == nil {
		return nil
	}
	return &model.Product{
		ID:            e.ID,
		Code:          e.Code,
		Name:          e.Name,
		Description:   e.Description,
		Price:         e.Price,
		Status:        e.Status,
		CreatedAt:     e.CreatedAt,
		CreatedByName: e.CreatedByName,
		UpdatedAt:     e.UpdatedAt,
		UpdatedByName: e.UpdatedByName,
	}
}

func ToEntityProduct(m *model.Product) *entity.Product {
	if m == nil {
		return nil
	}
	return &entity.Product{
		ID:            m.ID,
		Code:          m.Code,
		Name:          m.Name,
		Description:   m.Description,
		Price:         m.Price,
		Status:        m.Status,
		CreatedAt:     m.CreatedAt,
		CreatedByName: m.CreatedByName,
		UpdatedAt:     m.UpdatedAt,
		UpdatedByName: m.UpdatedByName,
	}
}

// ToEntityProducts always returns a non-nil slice (JSON "[]" not "null").
func ToEntityProducts(models []*model.Product) []*entity.Product {
	products := make([]*entity.Product, 0, len(models))
	for _, m := range models {
		products = append(products, ToEntityProduct(m))
	}
	return products
}
