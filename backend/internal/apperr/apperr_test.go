package apperr

import (
	"errors"
	"net/http"
	"testing"

	"backend-template/domain/errorcode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"repo.smf-indonesia.co.id/package-helper/go-common/pkg/apperror"
)

// Guards the rule "every errorcode has a catalog message".
func TestCatalog_EveryErrorCodeHasAMessage(t *testing.T) {
	for _, code := range errorcode.All {
		entry, ok := catalog[code]
		assert.Truef(t, ok, "errorcode %s has no entry in apperr catalog", code)
		assert.NotEmptyf(t, entry.ID, "errorcode %s has no Indonesian message", code)
		assert.NotEmptyf(t, entry.EN, "errorcode %s has no English message", code)
	}
}

func TestMessage(t *testing.T) {
	assert.Equal(t, "Produk tidak ditemukan", message(errorcode.ProductNotFound))
	assert.Equal(t, "Kode produk P-1 sudah digunakan", message(errorcode.ProductCodeAlreadyExists, "P-1"))
	assert.Equal(t, "UNKNOWN_CODE", message("UNKNOWN_CODE"), "unknown codes fall back to the code itself")
}

func TestConstructors_MapToHTTPStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"bad request", BadRequest(errorcode.InvalidRequest, "x"), http.StatusBadRequest},
		{"forbidden", Forbidden(errorcode.PermissionDenied), http.StatusForbidden},
		{"not found", NotFound(errorcode.ProductNotFound), http.StatusNotFound},
		{"conflict", Conflict(errorcode.ProductCodeAlreadyExists, "P-1"), http.StatusConflict},
		{"unprocessable", UnprocessableEntity(errorcode.InvalidRequest, "x"), http.StatusUnprocessableEntity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var appErr *apperror.AppError
			require.True(t, errors.As(tt.err, &appErr))
			assert.Equal(t, tt.want, appErr.Code)
		})
	}
}
