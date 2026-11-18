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
