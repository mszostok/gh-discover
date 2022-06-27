package xsignal

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// WithStopContext returns a copy of parent with a new Done channel. The returned
// context's Done channel is closed on of SIGINT or SIGTERM signals.
func WithStopContext(parent context.Context) context.Context {
	ctx, cancel := context.WithCancel(parent)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-ctx.Done():
		case <-sigCh:
			cancel()
		}
	}()

	return ctx
}
