// Package utils is the local re-implementation of go-common's utils package
// (constitution, Template Exception 4). Names and signatures match go-common
// v1.0.0; only the subset this service uses is implemented.
package utils

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"smart-grid-asset-registry/pkg/gocommon/pkg/apperror"
	"smart-grid-asset-registry/pkg/gocommon/pkg/logger"
	"smart-grid-asset-registry/pkg/gocommon/response"
)

const internalServerErrorMessage = "Internal Server Error"

// A request body above Fiber's BodyLimit never reaches a handler, so Fiber
// itself answers 413. This service maps that to the same application error
// code an over-size CSV gets from the usecase (spec FR-003, analysis N1), so
// every "file too large" answer carries a stable code. The values must stay in
// sync with errorcode.ImportFileTooLarge and its catalog message;
// internal/apperr's tests guard that.
const (
	PayloadTooLargeErrorCode = "IMPORT_FILE_TOO_LARGE"
	PayloadTooLargeMessage   = "The file is too large. The maximum size is 5 MB."
)

// ErrorData is the "data" of an error response that carries an application
// error code — the one addition to go-common (Template Exception 4).
type ErrorData struct {
	ErrorCode string `json:"errorCode"`
}

// NewErrorHandler is the global Fiber error handler:
//   - *apperror.AppError → its status and message; data.errorCode when the
//     error has an application error code, otherwise its Errors (usually null)
//   - *fiber.Error 413 → 413 with PayloadTooLargeMessage and
//     data.errorCode PayloadTooLargeErrorCode
//   - other *fiber.Error → its status and message
//   - anything else → 500 "Internal Server Error"; the real error is logged,
//     never sent to the client
//
// Errors answered with a status >= 500 are logged.
func NewErrorHandler(log *logrus.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			if appErr.Code >= fiber.StatusInternalServerError {
				logger.WithCtx(c.UserContext(), log).WithError(err).Error("app error")
			}
			if appErr.ErrorCode != "" {
				return response.Error(c, appErr.Code, appErr.Message, ErrorData{ErrorCode: appErr.ErrorCode})
			}
			return response.Error(c, appErr.Code, appErr.Message, appErr.Errors)
		}

		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			if fiberErr.Code == fiber.StatusRequestEntityTooLarge {
				return response.Error(c, fiberErr.Code, PayloadTooLargeMessage, ErrorData{ErrorCode: PayloadTooLargeErrorCode})
			}
			if fiberErr.Code >= fiber.StatusInternalServerError {
				logger.WithCtx(c.UserContext(), log).WithError(err).Error("fiber error")
			}
			return response.Error(c, fiberErr.Code, fiberErr.Message, nil)
		}

		logger.WithCtx(c.UserContext(), log).WithError(err).Error("unhandled error")
		return response.Error(c, fiber.StatusInternalServerError, internalServerErrorMessage, nil)
	}
}

// NotFoundHandler answers requests that matched no route.
func NotFoundHandler(c *fiber.Ctx) error {
	return response.Error(c, fiber.StatusNotFound, "Endpoint Not Found", nil)
}
