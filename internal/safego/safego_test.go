package safego_test

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/safego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGo_NoPanic(t *testing.T) {
	t.Parallel()

	var waitGroup sync.WaitGroup

	waitGroup.Add(1)

	safego.Go(context.Background(), func() {
		defer waitGroup.Done()
	})

	waitGroup.Wait()
}

func TestGo_PanicRecovered(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})

	safego.Go(context.Background(), func() {
		defer close(done)
		panic("test panic")
	})

	select {
	case <-done:
		// Goroutine completed — panic was recovered, not propagated.
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for panicking goroutine to complete")
	}
}

// recordingHandler captures slog records for test assertions.
// It is safe for concurrent use (slog handlers must be).
type recordingHandler struct {
	mu      sync.Mutex
	records []slog.Record
	done    chan struct{} // closed on first Handle call
}

func newRecordingHandler() *recordingHandler {
	return &recordingHandler{done: make(chan struct{})}
}

func (h *recordingHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *recordingHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.records = append(h.records, r)

	select {
	case <-h.done:
	default:
		close(h.done)
	}

	return nil
}

func (h *recordingHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *recordingHandler) WithGroup(_ string) slog.Handler      { return h }

func (h *recordingHandler) waitForRecord(t *testing.T) slog.Record {
	t.Helper()

	select {
	case <-h.done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for log record")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	require.NotEmpty(t, h.records, "expected at least one log record")

	return h.records[0]
}

// TestGo_PanicLogsError is intentionally non-parallel because it replaces
// the global slog default to capture log output.
func TestGo_PanicLogsError(t *testing.T) { //nolint:paralleltest // mutates global slog default
	handler := newRecordingHandler()
	logger := slog.New(handler)

	original := slog.Default()
	slog.SetDefault(logger)

	t.Cleanup(func() { slog.SetDefault(original) })

	safego.Go(context.Background(), func() {
		panic("kaboom")
	})

	record := handler.waitForRecord(t)

	assert.Equal(t, slog.LevelError, record.Level)
	assert.Equal(t, "panic recovered in goroutine", record.Message)

	// Verify the "panic" and "stack" attributes.
	var panicValue any

	var stackValue string

	record.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "panic":
			panicValue = a.Value.Any()
		case "stack":
			stackValue = a.Value.String()
		}

		return true
	})

	assert.Equal(t, "kaboom", panicValue)
	assert.NotEmpty(t, stackValue, "stack trace should be captured on panic")
	assert.Contains(t, stackValue, "safego.Go", "stack should reference safego.Go")
}
