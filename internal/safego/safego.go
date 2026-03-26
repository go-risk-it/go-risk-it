// Package safego provides a panic-recovering goroutine launcher.
// Use safego.Go instead of bare `go func()` for fire-and-forget goroutines
// to prevent panics from crashing the process.
package safego

import (
	"context"
	"log/slog"
	"runtime/debug"
)

// Go launches task in a new goroutine with panic recovery.
// If task panics, the panic is caught and logged via slog.ErrorContext
// with the recovered value. The panic does not propagate to the caller.
func Go(ctx context.Context, task func()) {
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.ErrorContext(
					ctx,
					"panic recovered in goroutine",
					"panic", recovered,
					"stack", string(debug.Stack()),
				)
			}
		}()

		task()
	}()
}
