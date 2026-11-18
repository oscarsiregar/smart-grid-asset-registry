package entity

import (
	"time"

	"backend-template/domain/valueobject"

	"github.com/google/uuid"
)

// Product is the domain representation of a row in the "product" table.
// Entities are plain structs: no gorm/json tags, no framework imports.
type Product struct {
	ID            uuid.UUID
	Code          string
	Name          string
	Description   *string
	Price         float64
	Status        valueobject.ProductStatus
	CreatedAt     time.Time
	CreatedByName string
	UpdatedAt     time.Time
	UpdatedByName string
}
