// Package apperr is the ONLY way business errors leave a usecase/handler.
// Each errorcode has an Indonesian message (sent to the client) and an
// English one (internal reference / future i18n). Never call go-common's
// apperror directly from usecases or handlers.
package apperr

import (
	"fmt"
	"net/http"

	"backend-template/domain/errorcode"

	"repo.smf-indonesia.co.id/package-helper/go-common/pkg/apperror"
)

type catalogEntry struct {
	ID string
	EN string
}

var catalog = map[string]catalogEntry{
	errorcode.InvalidRequest: {
		ID: "Permintaan tidak valid: %s",
		EN: "Invalid request: %s",
	},
	errorcode.PermissionDenied: {
		ID: "Anda tidak memiliki izin untuk mengakses data ini",
		EN: "You do not have permission to access this data",
	},
	errorcode.ProductNotFound: {
		ID: "Produk tidak ditemukan",
		EN: "Product not found",
	},
	errorcode.ProductCodeAlreadyExists: {
		ID: "Kode produk %s sudah digunakan",
		EN: "Product code %s is already in use",
	},
}

func message(code string, args ...interface{}) string {
	entry, ok := catalog[code]
	if !ok {
		return code
	}
	if len(args) == 0 {
		return entry.ID
	}
	return fmt.Sprintf(entry.ID, args...)
}

func BadRequest(code string, args ...interface{}) error {
	return apperror.BadRequest(message(code, args...))
}

func Forbidden(code string, args ...interface{}) error {
	return apperror.Forbidden(message(code, args...))
}

func NotFound(code string, args ...interface{}) error {
	return apperror.NotFound(message(code, args...))
}

func Conflict(code string, args ...interface{}) error {
	return apperror.Conflict(message(code, args...))
}

func UnprocessableEntity(code string, args ...interface{}) error {
	return apperror.New(http.StatusUnprocessableEntity, message(code, args...))
}
