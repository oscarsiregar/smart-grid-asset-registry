package apperr

import (
	"errors"
	"net/http"
	"testing"

	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/pkg/gocommon/pkg/apperror"

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

func TestMessage(t *testing.T) {
	assert.Equal(t, "Product not found", Message(errorcode.ProductNotFound))
	assert.Equal(t, "Product code P-1 is already in use", Message(errorcode.ProductCodeAlreadyExists, "P-1"))
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
		{"forbidden", Forbidden(errorcode.PermissionDenied), http.StatusForbidden, errorcode.PermissionDenied, "You do not have permission to access this data"},
		{"not found", NotFound(errorcode.ProductNotFound), http.StatusNotFound, errorcode.ProductNotFound, "Product not found"},
		{"conflict", Conflict(errorcode.ProductCodeAlreadyExists, "P-1"), http.StatusConflict, errorcode.ProductCodeAlreadyExists, "Product code P-1 is already in use"},
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
