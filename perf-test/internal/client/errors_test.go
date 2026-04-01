package client

import (
	"context"
	"io"
	"net"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// realTimeoutError simulates a genuine network timeout (e.g., dial timeout).
type realTimeoutError struct{ msg string }

func (e *realTimeoutError) Error() string   { return e.msg }
func (e *realTimeoutError) Timeout() bool   { return true }
func (e *realTimeoutError) Temporary() bool { return false }

// Compile-time proof that realTimeoutError is a net.Error.
var _ net.Error = (*realTimeoutError)(nil)

func TestClassifyNetError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		err         error
		wantNil     bool
		wantMessage string // substring check when non-nil
	}{
		{
			name:    "nil returns nil",
			err:     nil,
			wantNil: true,
		},
		{
			name:    "context.DeadlineExceeded returns nil (not retryable)",
			err:     context.DeadlineExceeded,
			wantNil: true,
		},
		{
			name:    "context.Canceled returns nil (not retryable)",
			err:     context.Canceled,
			wantNil: true,
		},
		{
			name:        "real network timeout returns TransientError",
			err:         &realTimeoutError{msg: "dial tcp: i/o timeout"},
			wantNil:     false,
			wantMessage: "transient network error",
		},
		{
			name:        "io.EOF returns TransientError",
			err:         io.EOF,
			wantNil:     false,
			wantMessage: "transient network error",
		},
		{
			name:        "ECONNRESET returns TransientError",
			err:         syscall.ECONNRESET,
			wantNil:     false,
			wantMessage: "transient network error",
		},
		{
			name:        "ECONNREFUSED returns TransientError",
			err:         syscall.ECONNREFUSED,
			wantNil:     false,
			wantMessage: "transient network error",
		},
		{
			name:    "unknown error returns nil",
			err:     assert.AnError,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := classifyNetError(tt.err)

			if tt.wantNil {
				assert.Nil(t, got)

				return
			}

			require.NotNil(t, got)

			var transient *TransientError
			require.ErrorAs(t, got, &transient)
			assert.Contains(t, transient.Error(), tt.wantMessage)
			assert.Equal(t, 0, transient.StatusCode)
		})
	}
}
