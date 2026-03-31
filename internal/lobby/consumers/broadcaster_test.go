package consumers_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/lobby/api/messaging"
	consumers "github.com/go-risk-it/go-risk-it/internal/lobby/consumers"
	"github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	lobbyevt "github.com/go-risk-it/go-risk-it/internal/lobby/events"
	"github.com/go-risk-it/go-risk-it/internal/lobby/logic/state"
	mockConsumers "github.com/go-risk-it/go-risk-it/mocks/internal_/lobby/consumers"
	mockState "github.com/go-risk-it/go-risk-it/mocks/internal_/lobby/logic/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace/noop"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const (
	testLobbyID = int64(7)
	testUserID  = "user-1"
)

func testLobbyContext() ctx.LobbyContext {
	return ctx.WithLobbyID(
		kernelctx.WithUserID(kernelctx.WithSpan(context.Background(), noop.Span{}), testUserID),
		testLobbyID,
	)
}

// testLobby returns a minimal valid *state.Lobby for the mock state.Service.
func testLobby() *state.Lobby {
	return &state.Lobby{
		ID: testLobbyID,
		Participants: []state.Participant{
			{UserID: "user-1"},
			{UserID: "user-2"},
		},
	}
}

// deps bundles all mock dependencies so each test can pick what it needs.
type deps struct {
	writer   *mockConsumers.Writer
	stateSvc *mockState.Service
}

func newDeps(t *testing.T) *deps {
	t.Helper()

	return &deps{
		writer:   mockConsumers.NewWriter(t),
		stateSvc: mockState.NewService(t),
	}
}

func (d *deps) newBroadcaster() *consumers.LobbyStateBroadcaster {
	stateCtrl := consumers.NewStateController(d.stateSvc)

	return consumers.NewLobbyStateBroadcaster(d.writer, stateCtrl)
}

// wsMessageType extracts the "type" field from a serialized WS message envelope.
func wsMessageType(t *testing.T, raw json.RawMessage) string {
	t.Helper()

	var envelope struct {
		Type string `json:"type"`
	}
	require.NoError(t, json.Unmarshal(raw, &envelope))

	return envelope.Type
}

// wsMessageData extracts the "data" payload from a serialized WS message envelope.
func wsMessageData(t *testing.T, raw json.RawMessage) json.RawMessage {
	t.Helper()

	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(raw, &envelope))

	return envelope.Data
}

// ---------------------------------------------------------------------------
// spyBus captures OnType registration calls for counting.
// ---------------------------------------------------------------------------

type spyBus struct {
	mu    sync.Mutex
	calls []string
}

var _ eventbus.Subscriber = (*spyBus)(nil)

func (s *spyBus) OnType(eventType string, _ eventbus.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls = append(s.calls, eventType)
}

func (s *spyBus) OnAll(eventbus.Handler) {}

func (s *spyBus) registrationCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.calls)
}

func (s *spyBus) registeredTypes() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]string, len(s.calls))
	copy(result, s.calls)

	return result
}

// ---------------------------------------------------------------------------
// Test: Registration count — exactly 2 OnLobbyEvent registrations
// ---------------------------------------------------------------------------

func TestRegister_CallsOnLobbyEventExactlyTwice(t *testing.T) {
	t.Parallel()

	d := newDeps(t)
	pub := d.newBroadcaster()

	spy := &spyBus{}
	pub.Register(spy)

	require.Equal(
		t,
		2,
		spy.registrationCount(),
		"Register should call OnLobbyEvent exactly 2 times",
	)

	types := spy.registeredTypes()
	assert.Contains(t, types, lobbyevt.TypeLobbyStateChanged)
	assert.Contains(t, types, lobbyevt.TypeLobbyPlayerConnected)
}

// ---------------------------------------------------------------------------
// Test: StateChanged broadcasts lobby state message
// ---------------------------------------------------------------------------

func TestOnStateChanged_BroadcastsLobbyState(t *testing.T) {
	t.Parallel()

	testDeps := newDeps(t)
	pub := testDeps.newBroadcaster()

	lobbyCtx := testLobbyContext()

	testDeps.stateSvc.EXPECT().
		GetLobbyState(mock.Anything).
		Return(testLobby(), nil)

	var captured json.RawMessage

	testDeps.writer.EXPECT().
		Broadcast(mock.Anything, mock.Anything).
		Run(func(_ ctx.LobbyContext, msg json.RawMessage) {
			captured = msg
		}).
		Return().
		Once()

	bus := eventbus.NewTestBus()
	pub.Register(bus)
	bus.Emit(lobbyCtx, lobbyevt.NewLobbyStateChanged(testLobbyID, testUserID))

	// Verify message type is "lobbyState".
	require.Equal(t, "lobbyState", wsMessageType(t, captured))

	// Verify payload contains expected lobby data.
	data := wsMessageData(t, captured)

	var lobbyState messaging.LobbyState
	require.NoError(t, json.Unmarshal(data, &lobbyState))
	assert.Equal(t, testLobbyID, lobbyState.ID)
	assert.Len(t, lobbyState.Participants, 2)
}

// ---------------------------------------------------------------------------
// Test: PlayerConnected sends lobby state via WriteMessage (not Broadcast)
// ---------------------------------------------------------------------------

func TestOnPlayerConnected_WritesLobbyState(t *testing.T) {
	t.Parallel()

	testDeps := newDeps(t)
	pub := testDeps.newBroadcaster()

	lobbyCtx := testLobbyContext()

	testDeps.stateSvc.EXPECT().
		GetLobbyState(mock.Anything).
		Return(testLobby(), nil)

	var captured json.RawMessage

	testDeps.writer.EXPECT().
		WriteMessage(mock.Anything, mock.Anything).
		Run(func(_ ctx.LobbyContext, msg json.RawMessage) {
			captured = msg
		}).
		Return().
		Once()

	bus := eventbus.NewTestBus()
	pub.Register(bus)
	bus.Emit(lobbyCtx, lobbyevt.NewLobbyPlayerConnected(testLobbyID, testUserID))

	// Verify message type is "lobbyState".
	require.Equal(t, "lobbyState", wsMessageType(t, captured))

	// Verify payload contains expected lobby data.
	data := wsMessageData(t, captured)

	var lobbyState messaging.LobbyState
	require.NoError(t, json.Unmarshal(data, &lobbyState))
	assert.Equal(t, testLobbyID, lobbyState.ID)
}

// ---------------------------------------------------------------------------
// Test: PlayerConnected does NOT call Broadcast
// ---------------------------------------------------------------------------

func TestOnPlayerConnected_NoBroadcastCalls(t *testing.T) {
	t.Parallel()

	testDeps := newDeps(t)
	pub := testDeps.newBroadcaster()

	lobbyCtx := testLobbyContext()

	testDeps.stateSvc.EXPECT().
		GetLobbyState(mock.Anything).
		Return(testLobby(), nil)

	// All messages go through WriteMessage, never Broadcast.
	testDeps.writer.EXPECT().
		WriteMessage(mock.Anything, mock.Anything).
		Return().
		Once()

	// Broadcast should NOT be called — if it is, the mock will fail
	// because there is no expectation set for it.

	bus := eventbus.NewTestBus()
	pub.Register(bus)
	bus.Emit(lobbyCtx, lobbyevt.NewLobbyPlayerConnected(testLobbyID, testUserID))
}

// ---------------------------------------------------------------------------
// Test: Panic recovery on StateChanged — panic in state fetch doesn't crash
// ---------------------------------------------------------------------------

func TestOnStateChanged_PanicInStateFetch_DoesNotCrash(t *testing.T) {
	t.Parallel()

	testDeps := newDeps(t)
	pub := testDeps.newBroadcaster()

	lobbyCtx := testLobbyContext()

	// State service panics — propagates through the real StateController
	// and is caught by safeOp in the broadcaster.
	testDeps.stateSvc.EXPECT().
		GetLobbyState(mock.Anything).
		Run(func(_ ctx.LobbyContext) {
			panic("simulated state service explosion")
		}).
		Return(nil, nil)

	// Broadcast should NOT be called since the fetch panicked.
	// No expectation set on writer.Broadcast — mock will fail if called.

	bus := eventbus.NewTestBus()
	pub.Register(bus)

	require.NotPanics(t, func() {
		bus.Emit(lobbyCtx, lobbyevt.NewLobbyStateChanged(testLobbyID, testUserID))
	})
}

// ---------------------------------------------------------------------------
// Test: Panic recovery on PlayerConnected — panic in state fetch doesn't crash
// ---------------------------------------------------------------------------

func TestOnPlayerConnected_PanicInStateFetch_DoesNotCrash(t *testing.T) {
	t.Parallel()

	d := newDeps(t)
	pub := d.newBroadcaster()

	lobbyCtx := testLobbyContext()

	d.stateSvc.EXPECT().
		GetLobbyState(mock.Anything).
		Run(func(_ ctx.LobbyContext) {
			panic("simulated state service explosion")
		}).
		Return(nil, nil)

	// WriteMessage should NOT be called since the fetch panicked.

	bus := eventbus.NewTestBus()
	pub.Register(bus)

	require.NotPanics(t, func() {
		bus.Emit(lobbyCtx, lobbyevt.NewLobbyPlayerConnected(testLobbyID, testUserID))
	})
}

// ---------------------------------------------------------------------------
// OTel setup helper — installs a real TracerProvider with InMemoryExporter,
// restores the previous global on cleanup.
// ---------------------------------------------------------------------------

func setupOTelExporter(t *testing.T) *tracetest.InMemoryExporter {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tracerProvider)

	t.Cleanup(func() {
		otel.SetTracerProvider(prev)
		_ = tracerProvider.Shutdown(context.Background())
	})

	return exporter
}

// spanNames extracts span names from stubs for diagnostic messages.
func spanNames(stubs tracetest.SpanStubs) []string {
	names := make([]string, len(stubs))
	for i, s := range stubs {
		names[i] = s.Name
	}

	return names
}

// ---------------------------------------------------------------------------
// Test: StateChanged — creates consumer.fetchLobbyState child span
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestOnStateChanged_CreatesSpan(t *testing.T) {
	// Not t.Parallel() — swaps global TracerProvider.
	exporter := setupOTelExporter(t)

	testDeps := newDeps(t)
	pub := testDeps.newBroadcaster()
	lobbyCtx := testLobbyContext()

	testDeps.stateSvc.EXPECT().
		GetLobbyState(mock.Anything).
		Return(testLobby(), nil)

	testDeps.writer.EXPECT().
		Broadcast(mock.Anything, mock.Anything).
		Return().
		Once()

	bus := eventbus.NewTestBus()
	pub.Register(bus)
	bus.Emit(lobbyCtx, lobbyevt.NewLobbyStateChanged(testLobbyID, testUserID))

	stubs := exporter.GetSpans()
	var foundFetch, foundDispatch bool

	for _, stub := range stubs {
		if stub.Name == "consumer.fetchLobbyState" {
			foundFetch = true
		}

		if stub.Name == "consumer.dispatchLobbyState" {
			foundDispatch = true
		}
	}

	require.True(t, foundFetch,
		"expected span named 'consumer.fetchLobbyState' in recorded spans, got: %v",
		spanNames(stubs))
	require.True(t, foundDispatch,
		"expected span named 'consumer.dispatchLobbyState' in recorded spans, got: %v",
		spanNames(stubs))
}
