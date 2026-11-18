package utils

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestShutdownHelper_Graceful(t *testing.T) {
	var logs bytes.Buffer
	log := logrus.New()
	log.SetOutput(&logs)

	var called []string
	NewShutdownHelper(log).Graceful(map[string]func(ctx context.Context) error{
		"Database": func(ctx context.Context) error {
			called = append(called, "Database")
			_, hasDeadline := ctx.Deadline()
			assert.True(t, hasDeadline)
			return nil
		},
		"Cache": func(context.Context) error {
			called = append(called, "Cache")
			return errors.New("close failed")
		},
	})

	assert.Equal(t, []string{"Cache", "Database"}, called)
	assert.Contains(t, logs.String(), "failed to shut down Cache")
	assert.Contains(t, logs.String(), "Database shut down")
}
