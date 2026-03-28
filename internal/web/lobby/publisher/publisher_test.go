package publisher_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/api/lobby/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/events"
	lobbyevt "github.com/go-risk-it/go-risk-it/internal/events/lobby"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/state"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/publisher"
	mockState "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/lobby/state"
	mockWS "github.com/go-risk-it/go-risk-it/mocks/internal_/web/lobby/ws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
		ctx.WithUserID(ctx.WithSpan(context.Background(), noop.Span{}), testUserID),
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
	writer   *mockWS.Writer
	stateSvc *mockState.Service
}

func newDeps(t *testing.T) *deps {
	t.Helper()

	return &deps{
		writer:   mockWS.NewWriter(t),
		stateSvc: mockState.NewService(t),
	}
}

func (d *deps) newPublisher() *publisher.LobbyStatePublisher {
	stateCtrl := controller.NewStateController(d.stateSvc)

	return publisher.NewLobbyStatePublisher(d.writer, stateCtrl)
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

var _ events.Bus = (*spyBus)(nil)

func (s *spyBus) OnType(eventType string, _ events.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls = append(s.calls, eventType)
}

func (s *spyBus) OnAll(events.Handler)               {}
func (s *spyBus) Emit(context.Context, events.Event) {}
func (s *spyBus) Close(context.Context) error        { return nil }

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
	pub := d.newPublisher()

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
	pub := testDeps.newPublisher()

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

	bus := events.NewTestBus()
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
	pub := testDeps.newPublisher()

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

	bus := events.NewTestBus()
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
	pub := testDeps.newPublisher()

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

	bus := events.NewTestBus()
	pub.Register(bus)
	bus.Emit(lobbyCtx, lobbyevt.NewLobbyPlayerConnected(testLobbyID, testUserID))
}

// ---------------------------------------------------------------------------
// Test: Panic recovery on StateChanged — panic in state fetch doesn't crash
// ---------------------------------------------------------------------------

func TestOnStateChanged_PanicInStateFetch_DoesNotCrash(t *testing.T) {
	t.Parallel()

	testDeps := newDeps(t)
	pub := testDeps.newPublisher()

	lobbyCtx := testLobbyContext()

	// State service panics — propagates through the real StateController
	// and is caught by safeOp in the publisher.
	testDeps.stateSvc.EXPECT().
		GetLobbyState(mock.Anything).
		Run(func(_ ctx.LobbyContext) {
			panic("simulated state service explosion")
		}).
		Return(nil, nil)

	// Broadcast should NOT be called since the fetch panicked.
	// No expectation set on writer.Broadcast — mock will fail if called.

	bus := events.NewTestBus()
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
	pub := d.newPublisher()

	lobbyCtx := testLobbyContext()

	d.stateSvc.EXPECT().
		GetLobbyState(mock.Anything).
		Run(func(_ ctx.LobbyContext) {
			panic("simulated state service explosion")
		}).
		Return(nil, nil)

	// WriteMessage should NOT be called since the fetch panicked.

	bus := events.NewTestBus()
	pub.Register(bus)

	require.NotPanics(t, func() {
		bus.Emit(lobbyCtx, lobbyevt.NewLobbyPlayerConnected(testLobbyID, testUserID))
	})
}
