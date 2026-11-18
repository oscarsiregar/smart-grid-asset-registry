package utils

import (
	"context"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

const defaultShutdownTimeout = 10 * time.Second

// ShutdownHelper waits for a termination signal and then closes resources.
type ShutdownHelper struct {
	log     *logrus.Logger
	timeout time.Duration
}

// NewShutdownHelper creates a helper with a 10-second timeout for Graceful.
func NewShutdownHelper(log *logrus.Logger) *ShutdownHelper {
	return &ShutdownHelper{log: log, timeout: defaultShutdownTimeout}
}

// Wait blocks until SIGINT or SIGTERM is received.
func (h *ShutdownHelper) Wait() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	received := <-quit
	h.log.Infof("shutdown signal received: %v", received)
}

// Graceful calls every shutdown function (in name order, for predictable
// logs) with a shared timeout and logs each result.
func (h *ShutdownHelper) Graceful(funcs map[string]func(ctx context.Context) error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	names := make([]string, 0, len(funcs))
	for name := range funcs {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		h.log.Infof("shutting down %s", name)
		if err := funcs[name](ctx); err != nil {
			h.log.WithError(err).Errorf("failed to shut down %s", name)
			continue
		}
		h.log.Infof("%s shut down", name)
	}
	h.log.Info("all resources shut down")
}
