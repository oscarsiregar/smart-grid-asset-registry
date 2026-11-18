// Package fiber is the local re-implementation of go-common's Fiber
// middleware package (constitution, Template Exception 4). Names and
// signatures match go-common v1.0.0; only the subset this service uses is
// implemented: request IDs, security headers and access logging. go-common's
// telemetry exporter and CORS setup are not needed behind the same-origin
// nginx proxy and are left out (see docs/KNOWN_LIMITATIONS.md).
package fiber

import (
	"context"
	"time"

	gofiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"smart-grid-asset-registry/pkg/gocommon/pkg/logger"
)

// InstallCommonMiddleware installs, in order: request/trace IDs, security
// headers and the access log.
func InstallCommonMiddleware(app *gofiber.App, log *logrus.Logger) {
	app.Use(requestID())
	app.Use(helmet.New())
	app.Use(accessLog(log))
}

// requestID reuses an incoming X-Request-ID or generates one, echoes it in the
// response, and puts it (also as the trace ID) into the request context so
// logger.WithCtx can add it to every log line.
func requestID() gofiber.Handler {
	return func(c *gofiber.Ctx) error {
		id := c.Get(gofiber.HeaderXRequestID)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(gofiber.HeaderXRequestID, id)

		ctx := context.WithValue(c.UserContext(), logger.RequestIDKey, id)
		ctx = context.WithValue(ctx, logger.TraceIDKey, id)
		c.SetUserContext(ctx)

		return c.Next()
	}
}

// accessLog writes one line per request. It logs the method, path, status and
// latency only — never headers or bodies.
func accessLog(log *logrus.Logger) gofiber.Handler {
	return func(c *gofiber.Ctx) error {
		start := time.Now()

		// Like Fiber's own logger middleware: hand a chain error to the
		// global error handler here, so the response (and the logged status)
		// is final, then report the request as handled.
		if chainErr := c.Next(); chainErr != nil {
			if handlerErr := c.App().Config().ErrorHandler(c, chainErr); handlerErr != nil {
				_ = c.SendStatus(gofiber.StatusInternalServerError)
			}
		}

		logger.WithCtx(c.UserContext(), log).WithFields(logrus.Fields{
			"method":     c.Method(),
			"path":       c.Path(),
			"status":     c.Response().StatusCode(),
			"latency_ms": time.Since(start).Milliseconds(),
		}).Info("request")

		return nil
	}
}
