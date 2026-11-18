package product_repository

import (
	"backend-template/domain/repository"

	"gorm.io/gorm"
)

// sortableColumns is the allow-list for ?sort= on the list endpoint.
var sortableColumns = []string{"code", "name", "price", "status", "created_at", "updated_at"}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) repository.ProductRepository {
	return &productRepository{
		db: db,
	}
}
