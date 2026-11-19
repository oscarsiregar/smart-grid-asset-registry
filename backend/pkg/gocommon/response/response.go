// Package response is the local re-implementation of go-common's response
// envelope (constitution, Template Exception 4). Names and signatures match
// go-common v1.0.0; only the subset this service uses is implemented.
package response

import (
	"github.com/gofiber/fiber/v2"
)

const (
	statusSuccess = "success"
	statusError   = "error"
)

// Response is the standard envelope. Data has no omitempty on purpose: the
// "data" key is always present, as null or as an empty list, so clients
// never have to guess whether it was left out.
type Response[T any] struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Meta    any    `json:"meta,omitempty"`
}

// Common is the non-generic form of Response.
type Common = Response[any]

// JSON writes the envelope with the given HTTP status. Status is "error" for
// codes >= 400 and "success" otherwise.
func JSON[T any](c *fiber.Ctx, code int, message string, data T) error {
	status := statusSuccess
	if code >= fiber.StatusBadRequest {
		status = statusError
	}

	return c.Status(code).JSON(Response[T]{
		Code:    code,
		Status:  status,
		Message: message,
		Data:    data,
	})
}

// Success writes a successful envelope.
func Success(c *fiber.Ctx, code int, message string, data interface{}) error {
	return JSON(c, code, message, data)
}

// Error writes an error envelope.
func Error(c *fiber.Ctx, code int, message string, data interface{}) error {
	return JSON(c, code, message, data)
}
