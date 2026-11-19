package logger

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	prod := New(true)
	assert.Equal(t, logrus.InfoLevel, prod.GetLevel())
	assert.IsType(t, &logrus.JSONFormatter{}, prod.Formatter)

	dev := New(false)
	assert.Equal(t, logrus.DebugLevel, dev.GetLevel())
	assert.IsType(t, &logrus.TextFormatter{}, dev.Formatter)
}

func TestWithCtx(t *testing.T) {
	log := New(false)

	tests := []struct {
		name string
		ctx  context.Context
		want logrus.Fields
	}{
		{name: "nil context", ctx: nil, want: logrus.Fields{}},
		{name: "no ids", ctx: context.Background(), want: logrus.Fields{}},
		{
			name: "request and trace id",
			ctx: context.WithValue(
				context.WithValue(context.Background(), RequestIDKey, "req-1"),
				TraceIDKey, "trace-1",
			),
			want: logrus.Fields{"request_id": "req-1", "trace_id": "trace-1"},
		},
		{name: "empty id ignored", ctx: context.WithValue(context.Background(), RequestIDKey, ""), want: logrus.Fields{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := WithCtx(tt.ctx, log)
			assert.Equal(t, tt.want, entry.Data)
		})
	}
}
