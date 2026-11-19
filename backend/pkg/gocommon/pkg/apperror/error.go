// Package apperror is the local re-implementation of go-common's apperror
// package (constitution, Template Exception 4). Names and signatures match
// go-common v1.0.0; only the subset this service uses is implemented.
package apperror

import "net/http"

// AppError is an error that carries the HTTP status it should be answered
// with. ErrorCode is the one addition to go-common (Template Exception 4):
// when set, the error handler sends it to the client as data.errorCode.
type AppError struct {
	Code      int
	Message   string
	Errors    interface{}
	ErrorCode string
}

func (e *AppError) Error() string {
	return e.Message
}

// New creates an AppError with the given HTTP status and message.
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func NotFound(message string) *AppError {
	return New(http.StatusNotFound, message)
}

func BadRequest(message string) *AppError {
	return New(http.StatusBadRequest, message)
}

func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, message)
}

func Conflict(message string) *AppError {
	return New(http.StatusConflict, message)
}
