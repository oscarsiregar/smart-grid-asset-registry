package apperr

import (
	"errors"
	"net/http"
	"testing"

	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/pkg/gocommon/pkg/apperror"
	"smart-grid-asset-registry/pkg/gocommon/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Guards the rule "every errorcode has a catalog message".
func TestCatalog_EveryErrorCodeHasAMessage(t *testing.T) {
	for _, code := range errorcode.All {
		message, ok := catalog[code]
		assert.Truef(t, ok, "errorcode %s has no entry in apperr catalog", code)
		assert.NotEmptyf(t, message, "errorcode %s has an empty message", code)
	}
}

// The go-common error handler answers Fiber's own 413 with a copy of this
// code and message (it may not import domain/); keep them identical.
func TestImportFileTooLarge_MatchesErrorHandler(t *testing.T) {
	assert.Equal(t, errorcode.ImportFileTooLarge, utils.PayloadTooLargeErrorCode)
	assert.Equal(t, Message(errorcode.ImportFileTooLarge, "5 MB"), utils.PayloadTooLargeMessage)
}

func TestMessage(t *testing.T) {
	assert.Equal(t, "Invalid request: x", Message(errorcode.InvalidRequest, "x"))
	assert.Equal(t, "Parent SS-9 is neither stored nor in the file.", Message(errorcode.ParentNotFound, "SS-9"))
	assert.Equal(t, "UNKNOWN_CODE", Message("UNKNOWN_CODE"), "unknown codes fall back to the code itself")
}

func TestConstructors_MapToHTTPStatusAndCarryTheCode(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{"bad request", BadRequest(errorcode.InvalidRequest, "x"), http.StatusBadRequest, errorcode.InvalidRequest, "Invalid request: x"},
		// The template's Forbidden/NotFound/Conflict stay available; this
		// feature has no code for them yet, so an unknown code (message =
		// the code itself) checks the status mapping.
		{"forbidden", Forbidden("SOME_CODE"), http.StatusForbidden, "SOME_CODE", "SOME_CODE"},
		{"not found", NotFound("SOME_CODE"), http.StatusNotFound, "SOME_CODE", "SOME_CODE"},
		{"conflict", Conflict(errorcode.AssetIDConflict, "TX-1", "voltage_kv"), http.StatusConflict, errorcode.AssetIDConflict,
			"Asset TX-1 is already stored with different values for: voltage_kv. Editing stored assets is not supported."},
		{"unprocessable", UnprocessableEntity(errorcode.InvalidRequest, "x"), http.StatusUnprocessableEntity, errorcode.InvalidRequest, "Invalid request: x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var appErr *apperror.AppError
			require.True(t, errors.As(tt.err, &appErr))
			assert.Equal(t, tt.wantStatus, appErr.Code)
			assert.Equal(t, tt.wantCode, appErr.ErrorCode)
			assert.Equal(t, tt.wantMessage, appErr.Message)
		})
	}
}
