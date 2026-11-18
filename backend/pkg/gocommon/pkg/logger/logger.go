// Package logger is the local re-implementation of go-common's logger
// package (constitution, Template Exception 4). Names and signatures match
// go-common v1.0.0; only the subset this service uses is implemented.
package logger

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

type contextKey string

const (
	// RequestIDKey and TraceIDKey are the context keys the common middleware
	// stores the request and trace IDs under.
	RequestIDKey contextKey = "request_id"
	TraceIDKey   contextKey = "trace_id"
)

// WithCtx returns a log entry carrying the request and trace IDs found in ctx.
func WithCtx(ctx context.Context, log *logrus.Logger) *logrus.Entry {
	fields := logrus.Fields{}
	if ctx == nil {
		return log.WithFields(fields)
	}

	if id, ok := ctx.Value(RequestIDKey).(string); ok && id != "" {
		fields["request_id"] = id
	}
	if id, ok := ctx.Value(TraceIDKey).(string); ok && id != "" {
		fields["trace_id"] = id
	}

	return log.WithFields(fields)
}

// New creates the service logger: JSON at Info level in production, readable
// text at Debug level elsewhere.
func New(isProd bool) *logrus.Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	if isProd {
		log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02T15:04:05.000Z07:00"})
		log.SetLevel(logrus.InfoLevel)
		return log
	}

	log.SetFormatter(&logrus.TextFormatter{TimestampFormat: "15:04:05.000", FullTimestamp: true})
	log.SetLevel(logrus.DebugLevel)
	return log
}
