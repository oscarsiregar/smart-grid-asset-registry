package entity

import (
	"time"

	"smart-grid-asset-registry/domain/valueobject"

	"github.com/google/uuid"
)

// AssetImport is the stored summary of one completed confirm attempt
// (spec FR-029). It holds counts only, never row details (FR-031).
//
// ImportedRows counts the new valid rows; they were saved only when Outcome
// is COMMITTED, so TotalRows = ImportedRows + UnchangedRows + RejectedRows
// always holds.
type AssetImport struct {
	ID            uuid.UUID
	FileName      string
	Mode          valueobject.ImportMode
	Outcome       valueobject.ImportOutcome
	TotalRows     int
	ImportedRows  int
	UnchangedRows int
	RejectedRows  int
	WarningRows   int
	CreatedAt     time.Time
	CreatedByName string
	UpdatedAt     time.Time
	UpdatedByName string
}
