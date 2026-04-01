package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
)

// TransientError wraps errors that are safe to retry (503, 429, timeouts, connection resets).
type TransientError struct {
	Cause      error
	StatusCode int // 0 for non-HTTP errors
}

func (e *TransientError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("transient HTTP %d: %v", e.StatusCode, e.Cause)
	}

	return fmt.Sprintf("transient network error: %v", e.Cause)
}

func (e *TransientError) Unwrap() error {
	return e.Cause
}

// StaleStateError is returned on HTTP 400 when the server rejects a move
// due to stale client state (e.g., "region does not have enough troops").
// The bot should wait for a fresh state update and re-decide, not retry
// the same invalid move.
type StaleStateError struct {
	Message string
}

func (e *StaleStateError) Error() string {
	return e.Message
}

// classifyHTTPStatus returns a TransientError for retryable HTTP status codes,
// or nil if the status is not retryable.
func classifyHTTPStatus(statusCode int, cause error) error {
	switch statusCode {
	case 429, 502, 503, 504:
		return &TransientError{Cause: cause, StatusCode: statusCode}
	default:
		return nil
	}
}

// classifyNetError returns a TransientError if the error is a retryable network
// error (timeout, connection reset/refused, EOF), or nil otherwise.
func classifyNetError(err error) error {
	if err == nil {
		return nil
	}

	// Context errors are fatal — retrying on a dead context is futile.
	// Must check BEFORE net.Error: context.DeadlineExceeded satisfies
	// net.Error (Timeout()=true), so the net.Error branch would catch it.
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return nil
	}

	// Connection dropped mid-read.
	if errors.Is(err, io.EOF) {
		return &TransientError{Cause: err}
	}

	// Net timeout (includes dial, TLS, read/write deadlines).
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return &TransientError{Cause: err}
	}

	// Syscall-level connection reset or refused.
	if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNREFUSED) {
		return &TransientError{Cause: err}
	}

	return nil
}
