// Package assetimport holds the pure CSV import rules (constitution, Template
// Exception 10): parsing and validating an import file against the stored
// assets it mentions. It imports only the standard library and other domain
// packages — no database, HTTP, framework or github.com/google/uuid — so every
// rule can be tested exhaustively without mocks.
package assetimport

import "smart-grid-asset-registry/domain/entity"

// Column is a CSV column name.
type Column string

// The eleven required CSV columns, in CSV order.
const (
	ColumnAssetID           Column = "asset_id"
	ColumnParentAssetID     Column = "parent_asset_id"
	ColumnAssetType         Column = "asset_type"
	ColumnAssetName         Column = "asset_name"
	ColumnOperationalStatus Column = "operational_status"
	ColumnVoltageKV         Column = "voltage_kv"
	ColumnRatingKVA         Column = "rating_kva"
	ColumnManufacturer      Column = "manufacturer"
	ColumnModel             Column = "model"
	ColumnSerialNumber      Column = "serial_number"
	ColumnCommissionedDate  Column = "commissioned_date"
)

// Columns lists every required column in CSV order.
var Columns = []Column{
	ColumnAssetID,
	ColumnParentAssetID,
	ColumnAssetType,
	ColumnAssetName,
	ColumnOperationalStatus,
	ColumnVoltageKV,
	ColumnRatingKVA,
	ColumnManufacturer,
	ColumnModel,
	ColumnSerialNumber,
	ColumnCommissionedDate,
}

// Row is one data record of the file.
type Row struct {
	// Number is the file line the record starts on; the header is line 1.
	Number int
	// Values holds the raw cell text per column, before trimming.
	Values map[Column]string
	// FieldCount is the number of fields the record had; a count different
	// from the header's makes the row MALFORMED_ROW.
	FieldCount int
}

// ParsedFile is a file whose header and CSV syntax are valid.
type ParsedFile struct {
	// HeaderFieldCount is the number of columns in the header.
	HeaderFieldCount int
	Rows             []Row
}

// FileError refuses the whole file (spec FR-003). Code is an errorcode
// constant; Args fill its catalog message.
type FileError struct {
	Code string
	Args []any
}

// StoredAssets are the already stored assets the file mentions.
type StoredAssets struct {
	// ByAssetID maps asset_id to the stored asset.
	ByAssetID map[string]*entity.Asset
	// BySerial maps a serial number to the asset IDs stored with it.
	BySerial map[string][]string
}

// Issue is one error or warning: a stable code plus the arguments for its
// catalog message. Text is rendered outside this package.
type Issue struct {
	Code string
	Args []any
}

// RejectedRow is a row that will not be imported, with every reason.
type RejectedRow struct {
	Row     int
	AssetID *string // nil when the row has no asset ID
	Errors  []Issue
}

// Warning is a non-blocking finding on an accepted row (spec FR-020).
type Warning struct {
	Row     int
	AssetID string
	Issue   Issue
}

// Counts summarises a report. Total = ToImport + Unchanged + Rejected.
type Counts struct {
	Total       int
	ToImport    int
	Unchanged   int
	Rejected    int
	WarningRows int
}

// Report is the result of validating a file.
type Report struct {
	// Accepted are the new assets to insert, parents first (depth, then
	// asset ID). ID, ImportID and the audit fields are left zero; the
	// usecase assigns them.
	Accepted      []*entity.Asset
	UnchangedRows []int
	Rejected      []RejectedRow // sorted by row
	Warnings      []Warning     // sorted by row
	Counts        Counts
}
