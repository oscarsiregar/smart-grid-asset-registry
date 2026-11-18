package errorcode

// Error codes are stable, SCREAMING_SNAKE_CASE identifiers. Every code here
// MUST have a matching English message in internal/apperr/apperr.go —
// TestCatalog_EveryErrorCodeHasAMessage enforces that.
const (
	InvalidRequest   = "INVALID_REQUEST"
	PermissionDenied = "PERMISSION_DENIED"

	ProductNotFound          = "PRODUCT_NOT_FOUND"
	ProductCodeAlreadyExists = "PRODUCT_CODE_ALREADY_EXISTS"
)

// CSV import: file-level codes. The whole file is refused and no row results
// are returned (spec FR-003).
const (
	ImportFileTooLarge     = "IMPORT_FILE_TOO_LARGE"
	ImportFileEmpty        = "IMPORT_FILE_EMPTY"
	ImportFileNotCSV       = "IMPORT_FILE_NOT_CSV"
	ImportMissingColumns   = "IMPORT_MISSING_COLUMNS"
	ImportDuplicateColumns = "IMPORT_DUPLICATE_COLUMNS"
)

// CSV import: row-level codes. The row is rejected; a row may carry several
// (spec FR-005 – FR-019, data-model.md).
const (
	MalformedRow             = "MALFORMED_ROW"
	AssetIDRequired          = "ASSET_ID_REQUIRED"
	AssetTypeRequired        = "ASSET_TYPE_REQUIRED"
	AssetNameRequired        = "ASSET_NAME_REQUIRED"
	DuplicateAssetID         = "DUPLICATE_ASSET_ID"
	InvalidAssetType         = "INVALID_ASSET_TYPE"
	InvalidOperationalStatus = "INVALID_OPERATIONAL_STATUS"
	InvalidNumber            = "INVALID_NUMBER"
	NegativeRating           = "NEGATIVE_RATING"
	InvalidDate              = "INVALID_DATE"
	SubstationHasParent      = "SUBSTATION_HAS_PARENT"
	ParentRequired           = "PARENT_REQUIRED"
	SelfParent               = "SELF_PARENT"
	ParentNotFound           = "PARENT_NOT_FOUND"
	InvalidParentType        = "INVALID_PARENT_TYPE"
	CycleDetected            = "CYCLE_DETECTED"
	ParentRejected           = "PARENT_REJECTED"
	AssetIDConflict          = "ASSET_ID_CONFLICT"
)

// CSV import: warning codes. The row is still accepted (spec FR-020).
const (
	DuplicateSerialNumber = "DUPLICATE_SERIAL_NUMBER"
)

// All lists every code above. Add new codes here too.
var All = []string{
	InvalidRequest,
	PermissionDenied,
	ProductNotFound,
	ProductCodeAlreadyExists,

	ImportFileTooLarge,
	ImportFileEmpty,
	ImportFileNotCSV,
	ImportMissingColumns,
	ImportDuplicateColumns,

	MalformedRow,
	AssetIDRequired,
	AssetTypeRequired,
	AssetNameRequired,
	DuplicateAssetID,
	InvalidAssetType,
	InvalidOperationalStatus,
	InvalidNumber,
	NegativeRating,
	InvalidDate,
	SubstationHasParent,
	ParentRequired,
	SelfParent,
	ParentNotFound,
	InvalidParentType,
	CycleDetected,
	ParentRejected,
	AssetIDConflict,

	DuplicateSerialNumber,
}
