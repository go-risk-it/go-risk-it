package observe_test

import (
	"errors"
	"testing"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	lobbyctx "github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
)

// ---------------------------------------------------------------------------
// Helpers (shared with observe_test.go)
// ---------------------------------------------------------------------------

// buildLobbyContext creates a full LobbyContext chain for testing context
// attribute extraction and rebase behavior.
func buildLobbyContext(t *testing.T) lobbyctx.LobbyContext {
	t.Helper()

	parentCtx, parentSpan := startParentSpan(t)
	traceCtx := ctx.WithSpan(parentCtx, parentSpan)
	userCtx := ctx.WithUserID(traceCtx, "user-42")

	return lobbyctx.WithLobbyID(userCtx, 77)
}

// ---------------------------------------------------------------------------
// Tests: Span (closure-based, returns (T, error))
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestSpan_GameContext_PreservesType(t *testing.T) {
	setupTracing(t)

	gameCtx := buildGameContext(t)

	result, err := observe.Span(
		gameCtx,
		"game.typed-func",
		func(gc gamectx.GameContext) (int64, error) {
			// fn receives GameContext directly — no assertion needed.
			return gc.GameID(), nil
		},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(99), result)
}

//nolint:paralleltest // swaps global TracerProvider
func TestSpan_LobbyContext_PreservesType(t *testing.T) {
	setupTracing(t)

	lCtx := buildLobbyContext(t)

	result, err := observe.Span(
		lCtx,
		"lobby.typed-func",
		func(lc lobbyctx.LobbyContext) (int64, error) {
			return lc.LobbyID(), nil
		},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(77), result)
}

//nolint:paralleltest // swaps global TracerProvider
func TestSpan_PreservesTraceChain(t *testing.T) {
	exporter := setupTracing(t)

	gameCtx := buildGameContext(t)
	parentSpan := gameCtx.Span()
	parentTraceID := parentSpan.SpanContext().TraceID()
	parentSpanID := parentSpan.SpanContext().SpanID()

	_, err := observe.Span(
		gameCtx,
		"game.trace-chain",
		func(_ gamectx.GameContext) (struct{}, error) {
			return struct{}{}, nil
		},
	)
	require.NoError(t, err)

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "game.trace-chain")
	require.NotNil(t, stub, "child span must be recorded")

	assert.Equal(t, parentTraceID, stub.SpanContext.TraceID(),
		"child must be in same trace as parent")
	assert.Equal(t, parentSpanID, stub.Parent.SpanID(),
		"child's parent must be the parent span")
}

//nolint:paralleltest // swaps global TracerProvider
func TestSpan_ErrorRecordedOnSpan(t *testing.T) {
	exporter := setupTracing(t)

	gameCtx := buildGameContext(t)
	testErr := errors.New("validation failed")

	_, err := observe.Span(
		gameCtx,
		"game.failing-op",
		func(_ gamectx.GameContext) (struct{}, error) {
			return struct{}{}, testErr
		},
	)
	require.Error(t, err)

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "game.failing-op")
	require.NotNil(t, stub, "span must be recorded")

	assert.Equal(t, codes.Error, stub.Status.Code, "span must have Error status")
	assert.Equal(t, "validation failed", stub.Status.Description)

	errEvt := findEvent(stub, "exception")
	require.NotNil(t, errEvt, "span must have an exception event from RecordError")
}

// ---------------------------------------------------------------------------
// Tests: SpanErr (closure-based, returns error)
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestSpanErr_GameContext(t *testing.T) {
	setupTracing(t)

	gameCtx := buildGameContext(t)

	var captured int64

	err := observe.SpanErr(
		gameCtx,
		"game.typed-err",
		func(gc gamectx.GameContext) error {
			captured = gc.GameID()

			return nil
		},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(99), captured)
}
