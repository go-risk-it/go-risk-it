package observe_test

import (
	"context"
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
// Tests: TypedSpan
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestTypedSpan_GameContext_ReturnsGameContext(t *testing.T) {
	setupTracing(t)

	gameCtx := buildGameContext(t)

	result, done := observe.TypedSpan(gameCtx, "game.test")
	defer done(nil)

	// The returned context must be a GameContext with the same GameID.
	require.NotNil(t, result)
	assert.Equal(t, int64(99), result.GameID())
	assert.Equal(t, "user-42", result.UserID())
}

//nolint:paralleltest // swaps global TracerProvider
func TestTypedSpan_LobbyContext_ReturnsLobbyContext(t *testing.T) {
	setupTracing(t)

	lobbyCtx := buildLobbyContext(t)

	result, done := observe.TypedSpan(lobbyCtx, "lobby.test")
	defer done(nil)

	// The returned context must be a LobbyContext with the same LobbyID.
	require.NotNil(t, result)
	assert.Equal(t, int64(77), result.LobbyID())
	assert.Equal(t, "user-42", result.UserID())
}

//nolint:paralleltest // swaps global TracerProvider
func TestTypedSpan_CreatesChildSpan(t *testing.T) {
	exporter := setupTracing(t)

	gameCtx := buildGameContext(t)
	parentSpan := gameCtx.Span()
	parentTraceID := parentSpan.SpanContext().TraceID()
	parentSpanID := parentSpan.SpanContext().SpanID()

	_, done := observe.TypedSpan(gameCtx, "game.child-op")
	done(nil)

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "game.child-op")
	require.NotNil(t, stub, "child span must be recorded")

	assert.Equal(t, parentTraceID, stub.SpanContext.TraceID(),
		"child must be in same trace as parent")
	assert.Equal(t, parentSpanID, stub.Parent.SpanID(),
		"child's parent must be the parent span")
}

//nolint:paralleltest // swaps global TracerProvider
func TestTypedSpan_DoneRecordsError(t *testing.T) {
	exporter := setupTracing(t)

	gameCtx := buildGameContext(t)

	_, done := observe.TypedSpan(gameCtx, "game.failing-op")

	testErr := errors.New("validation failed")
	done(testErr)

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "game.failing-op")
	require.NotNil(t, stub, "span must be recorded")

	assert.Equal(t, codes.Error, stub.Status.Code, "span must have Error status")
	assert.Equal(t, "validation failed", stub.Status.Description)

	errEvt := findEvent(stub, "exception")
	require.NotNil(t, errEvt, "span must have an exception event from RecordError")
}

// ---------------------------------------------------------------------------
// Tests: TypedSpanFunc
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestTypedSpanFunc_GameContext(t *testing.T) {
	setupTracing(t)

	gameCtx := buildGameContext(t)

	result, err := observe.TypedSpanFunc(
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
func TestTypedSpanErr_GameContext(t *testing.T) {
	setupTracing(t)

	gameCtx := buildGameContext(t)

	var captured int64

	err := observe.TypedSpanErr(
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

//nolint:paralleltest // swaps global TracerProvider
func TestTypedSpanFunc_PreservesTraceChain(t *testing.T) {
	exporter := setupTracing(t)

	gameCtx := buildGameContext(t)
	parentSpan := gameCtx.Span()
	parentTraceID := parentSpan.SpanContext().TraceID()
	parentSpanID := parentSpan.SpanContext().SpanID()

	_, err := observe.TypedSpanFunc(
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

// ---------------------------------------------------------------------------
// Tests: SpanFunc
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestSpanFunc_Success(t *testing.T) {
	exporter := setupTracing(t)

	parentCtx, _ := startParentSpan(t)

	result, err := observe.SpanFunc(parentCtx, "compute.value",
		func(_ context.Context) (int64, error) {
			return 42, nil
		},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(42), result)

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "compute.value")
	require.NotNil(t, stub, "span must be recorded")
	assert.Equal(t, codes.Unset, stub.Status.Code, "span must not have error status on nil")
}

//nolint:paralleltest // swaps global TracerProvider
func TestSpanFunc_Error(t *testing.T) {
	exporter := setupTracing(t)

	parentCtx, _ := startParentSpan(t)

	testErr := errors.New("db timeout")

	result, err := observe.SpanFunc(parentCtx, "compute.failing",
		func(_ context.Context) (int64, error) {
			return 0, testErr
		},
	)

	require.ErrorIs(t, err, testErr)
	assert.Equal(t, int64(0), result)

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "compute.failing")
	require.NotNil(t, stub, "span must be recorded")
	assert.Equal(t, codes.Error, stub.Status.Code, "span must have Error status")
	assert.Equal(t, "db timeout", stub.Status.Description)

	errEvt := findEvent(stub, "exception")
	require.NotNil(t, errEvt, "span must have an exception event from RecordError")
}

//nolint:paralleltest // swaps global TracerProvider
func TestSpanFunc_WithGameContext(t *testing.T) {
	setupTracing(t)

	gameCtx := buildGameContext(t)

	var insideGameID int64

	_, err := observe.SpanFunc(gameCtx, "game.create",
		func(innerCtx context.Context) (int64, error) {
			// The inner context should be a GameContext (auto-rebased by Span).
			gc, ok := innerCtx.(gamectx.GameContext)
			require.True(t, ok, "inner context must be GameContext after auto-rebase")
			insideGameID = gc.GameID()

			return 1, nil
		},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(99), insideGameID,
		"GameID must be preserved through SpanFunc auto-rebase")
}

// ---------------------------------------------------------------------------
// Tests: SpanErr
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestSpanErr_Success(t *testing.T) {
	exporter := setupTracing(t)

	parentCtx, _ := startParentSpan(t)

	err := observe.SpanErr(parentCtx, "validate.input",
		func(_ context.Context) error {
			return nil
		},
	)

	require.NoError(t, err)

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "validate.input")
	require.NotNil(t, stub, "span must be recorded")
	assert.Equal(t, codes.Unset, stub.Status.Code, "span must not have error status on nil")
}

//nolint:paralleltest // swaps global TracerProvider
func TestSpanErr_Error(t *testing.T) {
	exporter := setupTracing(t)

	parentCtx, _ := startParentSpan(t)

	testErr := errors.New("invalid input")

	err := observe.SpanErr(parentCtx, "validate.failing",
		func(_ context.Context) error {
			return testErr
		},
	)

	require.ErrorIs(t, err, testErr)

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "validate.failing")
	require.NotNil(t, stub, "span must be recorded")
	assert.Equal(t, codes.Error, stub.Status.Code, "span must have Error status")
	assert.Equal(t, "invalid input", stub.Status.Description)

	errEvt := findEvent(stub, "exception")
	require.NotNil(t, errEvt, "span must have an exception event from RecordError")
}
