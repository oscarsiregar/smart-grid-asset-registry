package model

import (
	"time"

	"smart-grid-asset-registry/domain/valueobject"

	"github.com/google/uuid"
)

// Asset is the GORM model for table "asset". Models carry gorm tags and
// TableName(); they never leave internal/repository/psql (use mapper/).
//
// Audit columns are *ByName on purpose (rule 05): a field literally named
// CreatedBy/UpdatedBy would be managed by go-common's audit plugin.
type Asset struct {
	ID                uuid.UUID                      `gorm:"column:id;primaryKey"`
	AssetID           string                         `gorm:"column:asset_id"`
	ParentAssetID     *string                        `gorm:"column:parent_asset_id"`
	AssetType         valueobject.AssetType          `gorm:"column:asset_type"`
	AssetName         string                         `gorm:"column:asset_name"`
	OperationalStatus *valueobject.OperationalStatus `gorm:"column:operational_status"`
	VoltageKV         *float64                       `gorm:"column:voltage_kv"`
	RatingKVA         *float64                       `gorm:"column:rating_kva"`
	Manufacturer      *string                        `gorm:"column:manufacturer"`
	Model             *string                        `gorm:"column:model"`
	SerialNumber      *string                        `gorm:"column:serial_number"`
	CommissionedDate  *time.Time                     `gorm:"column:commissioned_date;type:date"`
	ImportID          uuid.UUID                      `gorm:"column:import_id"`
	CreatedAt         time.Time                      `gorm:"column:created_at"`
	CreatedByName     string                         `gorm:"column:created_by_name"`
	UpdatedAt         time.Time                      `gorm:"column:updated_at"`
	UpdatedByName     string                         `gorm:"column:updated_by_name"`
}

func (Asset) TableName() string {
	return "asset"
}
