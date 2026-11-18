// Package apperr is the ONLY way business errors leave a usecase/handler.
// Each errorcode has one English message (constitution, Template Exception 6)
// that is sent to the client. Every error built here also carries its code,
// which the error handler sends as data.errorCode (Template Exception 4).
// Never call go-common's apperror directly from usecases or handlers.
package apperr

import (
	"fmt"
	"net/http"

	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/pkg/gocommon/pkg/apperror"
)

var catalog = map[string]string{
	errorcode.InvalidRequest:           "Invalid request: %s",
	errorcode.PermissionDenied:         "You do not have permission to access this data",
	errorcode.ProductNotFound:          "Product not found",
	errorcode.ProductCodeAlreadyExists: "Product code %s is already in use",

	// CSV import, file level. Args: limit; –; line; column list; column list.
	errorcode.ImportFileTooLarge:     "The file is too large. The maximum size is %s.",
	errorcode.ImportFileEmpty:        "The file contains no data rows.",
	errorcode.ImportFileNotCSV:       "The file is not valid UTF-8 CSV text (problem on line %d).",
	errorcode.ImportMissingColumns:   "The file is missing required columns: %s.",
	errorcode.ImportDuplicateColumns: "The file header repeats columns: %s.",

	// CSV import, row level. Args are listed in the order the format uses them.
	errorcode.MalformedRow:             "The row has %d fields, but the header has %d.",
	errorcode.AssetIDRequired:          "asset_id is required.",
	errorcode.AssetTypeRequired:        "asset_type is required.",
	errorcode.AssetNameRequired:        "asset_name is required.",
	errorcode.DuplicateAssetID:         "Asset ID %s appears on more than one row (rows %s).",
	errorcode.InvalidAssetType:         "Asset type %s is not supported. Use SUBSTATION, TRANSFORMER, LV_BOARD, SWITCHBOARD or SWITCHBOARD_PANEL.",
	errorcode.InvalidOperationalStatus: "Operational status %s is not supported. Use IN_SERVICE, MAINTENANCE or OUT_OF_SERVICE.",
	errorcode.InvalidNumber:            "%s must be a number, but is %s.",
	errorcode.NegativeRating:           "rating_kva must not be negative, but is %s.",
	errorcode.InvalidDate:              "commissioned_date %s is not a valid date. Use d/m/yy (for example 15/2/09) or yyyy-mm-dd.",
	errorcode.SubstationHasParent:      "A SUBSTATION cannot have a parent, but its parent is %s.",
	errorcode.ParentRequired:           "A %s must have a parent asset.",
	errorcode.SelfParent:               "An asset cannot be its own parent.",
	errorcode.ParentNotFound:           "Parent %s is neither stored nor in the file.",
	errorcode.InvalidParentType:        "A %s must have a %s as parent, but %s is a %s.",
	errorcode.CycleDetected:            "%s is part of a parent cycle: %s.",
	errorcode.ParentRejected:           "Parent %s is rejected, so this row cannot be imported.",
	errorcode.AssetIDConflict:          "Asset %s is already stored with different values for: %s. Editing stored assets is not supported.",

	// CSV import, warnings. Args: serial number, other asset IDs.
	errorcode.DuplicateSerialNumber: "Serial number %s is also used by %s.",
}

// Message renders the English message for code with args. Unknown codes fall
// back to the code itself so a missing entry is visible rather than silent.
func Message(code string, args ...any) string {
	template, ok := catalog[code]
	if !ok {
		return code
	}
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}

func newError(status int, code string, args ...any) error {
	appErr := apperror.New(status, Message(code, args...))
	appErr.ErrorCode = code
	return appErr
}

func BadRequest(code string, args ...any) error {
	return newError(http.StatusBadRequest, code, args...)
}

func Forbidden(code string, args ...any) error {
	return newError(http.StatusForbidden, code, args...)
}

func NotFound(code string, args ...any) error {
	return newError(http.StatusNotFound, code, args...)
}

func Conflict(code string, args ...any) error {
	return newError(http.StatusConflict, code, args...)
}

func UnprocessableEntity(code string, args ...any) error {
	return newError(http.StatusUnprocessableEntity, code, args...)
}
