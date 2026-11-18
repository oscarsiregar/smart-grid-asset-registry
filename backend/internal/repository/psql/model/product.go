package model

import (
	"time"

	"smart-grid-asset-registry/domain/valueobject"

	"github.com/google/uuid"
)

// Product is the GORM model for table "product". Models carry gorm tags and
// TableName(); they never leave internal/repository/psql (use mapper/).
//
// Audit columns are *ByName on purpose: go-common's audit plugin (installed
// by database.Connect) hijacks any field literally named CreatedBy/UpdatedBy
// — it overwrites it with the JWT user UUID, and silently DROPS the value
// when it is not a UUID. Only name a field CreatedBy/UpdatedBy if you want
// exactly that behavior (UUID column, see go-common database.AuditModel).
type Product struct {
	ID            uuid.UUID                 `gorm:"column:id;primaryKey"`
	Code          string                    `gorm:"column:code"`
	Name          string                    `gorm:"column:name"`
	Description   *string                   `gorm:"column:description"`
	Price         float64                   `gorm:"column:price"`
	Status        valueobject.ProductStatus `gorm:"column:status"`
	CreatedAt     time.Time                 `gorm:"column:created_at"`
	CreatedByName string                    `gorm:"column:created_by_name"`
	UpdatedAt     time.Time                 `gorm:"column:updated_at"`
	UpdatedByName string                    `gorm:"column:updated_by_name"`
}

func (Product) TableName() string {
	return "product"
}
