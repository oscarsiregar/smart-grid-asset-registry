package model

import (
	"time"

	"smart-grid-asset-registry/domain/valueobject"

	"github.com/google/uuid"
)

// AssetImport is the GORM model for table "asset_import" (see model.Asset
// for the conventions).
type AssetImport struct {
	ID            uuid.UUID                 `gorm:"column:id;primaryKey"`
	FileName      string                    `gorm:"column:file_name"`
	Mode          valueobject.ImportMode    `gorm:"column:mode"`
	Outcome       valueobject.ImportOutcome `gorm:"column:outcome"`
	TotalRows     int                       `gorm:"column:total_rows"`
	ImportedRows  int                       `gorm:"column:imported_rows"`
	UnchangedRows int                       `gorm:"column:unchanged_rows"`
	RejectedRows  int                       `gorm:"column:rejected_rows"`
	WarningRows   int                       `gorm:"column:warning_rows"`
	CreatedAt     time.Time                 `gorm:"column:created_at"`
	CreatedByName string                    `gorm:"column:created_by_name"`
	UpdatedAt     time.Time                 `gorm:"column:updated_at"`
	UpdatedByName string                    `gorm:"column:updated_by_name"`
}

func (AssetImport) TableName() string {
	return "asset_import"
}
