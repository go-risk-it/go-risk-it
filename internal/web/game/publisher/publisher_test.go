package publisher_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/events"
	gameevt "github.com/go-risk-it/go-risk-it/internal/events/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/game/publisher"
	mockMission "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/mission"
	mockLogging "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/move/orchestration/logging"
	mockSnapshot "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/snapshot"
	mockWS "github.com/go-risk-it/go-risk-it/mocks/internal_/web/game/ws"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const (
	testGameID  = int64(42)
	testUserID  = "user-1"
	testUserID2 = "user-2"
	historySize = int64(50)
)

func testGameContext() ctx.GameContext {
	return ctx.WithGameID(
		ctx.WithUserID(ctx.WithSpan(context.Background(), noop.Span{}), testUserID),
		testGameID,
	)
}

// testPublicSnapshot returns a minimal valid PublicSnapshot that passes through
// the converter pipeline without error (ATTACK phase needs no extra state).
func testPublicSnapshot() *snapshot.PublicSnapshot {
	return &snapshot.PublicSnapshot{
		Game: sqlc.GetGameRow{
			ID:           testGameID,
			CurrentPhase: sqlc.GamePhaseTypeATTACK,
			Turn:         3,
		},
		Phase: snapshot.PhaseState{Type: sqlc.GamePhaseTypeATTACK},
		Board: []sqlc.GetRegionsByGameRow{
			{ID: 1, ExternalReference: "alaska", Troops: 5, UserID: testUserID},
		},
		Players: []sqlc.GetPlayersStateRow{
			{UserID: testUserID, Name: "Alice", TurnIndex: 0, CardCount: 2, RegionCount: 10},
			{UserID: testUserID2, Name: "Bob", TurnIndex: 1, CardCount: 1, RegionCount: 8},
		},
	}
}

// testPrivateSnapshots returns per-user private snapshots keyed by user ID.
// Uses a static mission type (TWENTY_FOUR_TERRITORIES) so no mission service
// calls are needed.
func testPrivateSnapshots() map[string]*snapshot.PrivateSnapshot {
	return map[string]*snapshot.PrivateSnapshot{
		testUserID: {
			Cards:       nil,
			MissionType: sqlc.GameMissionTypeTWENTYFOURTERRITORIES,
			MissionID:   1,
		},
		testUserID2: {
			Cards:       nil,
			MissionType: sqlc.GameMissionTypeTWENTYFOURTERRITORIES,
			MissionID:   2,
		},
	}
}

// testMoveExecutedEvent returns a minimal MoveExecuted event.
func testMoveExecutedEvent() *gameevt.MoveExecuted {
	return gameevt.NewMoveExecuted(
		testGameID,
		testUserID,
		time.Now(),
		sqlc.GamePhaseTypeATTACK,
		sqlc.GameMoveLog{
			ID:       99,
			GameID:   testGameID,
			PlayerID: 1,
			Phase:    sqlc.GamePhaseTypeATTACK,
			MoveData: json.RawMessage(`{}`),
			Result:   json.RawMessage(`{}`),
			Created:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
		},
		sqlc.GamePhaseTypeATTACK,
		false,
		3,
		nil,
		nil,
	)
}

// deps bundles all mock dependencies so each test can pick what it needs.
type deps struct {
	writer     *mockWS.Writer
	presence   *mockWS.Presence
	lifecycle  *mockWS.Lifecycle
	snapSvc    *mockSnapshot.Service
	missionSvc *mockMission.Service
	loggingSvc *mockLogging.Service
}

func newDeps(t *testing.T) *deps {
	t.Helper()

	return &deps{
		writer:     mockWS.NewWriter(t),
		presence:   mockWS.NewPresence(t),
		lifecycle:  mockWS.NewLifecycle(t),
		snapSvc:    mockSnapshot.NewService(t),
		missionSvc: mockMission.NewService(t),
		loggingSvc: mockLogging.NewService(t),
	}
}

func (d *deps) newPublisher() *publisher.GameStatePublisher {
	return publisher.NewGameStatePublisher(
		d.writer,
		d.presence,
		d.lifecycle,
		d.snapSvc,
		controller.NewMissionController(d.missionSvc),
		controller.NewMoveLogController(d.loggingSvc),
		config.HistoryConfig{Size: historySize},
	)
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

// ---------------------------------------------------------------------------
// spyBus — captures OnType registration calls for counting.
// ---------------------------------------------------------------------------

type spyBus struct {
	mu    sync.Mutex
	calls []string // event types registered via OnType
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
// Test: Registration count
// ---------------------------------------------------------------------------

func TestRegister_CallsOnGameEventExactlyThreeTimes(t *testing.T) {
	t.Parallel()

	d := newDeps(t)
	pub := d.newPublisher()

	spy := &spyBus{}
	pub.Register(spy)

	require.Equal(
		t,
		3,
		spy.registrationCount(),
		"Register should call OnGameEvent exactly 3 times",
	)

	types := spy.registeredTypes()
	assert.Contains(t, types, gameevt.TypeMoveExecuted)
	assert.Contains(t, types, gameevt.TypeGameCompleted)
	assert.Contains(t, types, gameevt.TypePlayerConnected)
}

// ---------------------------------------------------------------------------
// Test: MoveExecuted sequential ordering
// ---------------------------------------------------------------------------

func TestHandleMoveExecuted_BroadcastOrdering(t *testing.T) {
	t.Parallel()

	testDeps := newDeps(t)
	pub := testDeps.newPublisher()
	gameCtx := testGameContext()
	event := testMoveExecutedEvent()

	// --- setup snapshot mocks ---
	testDeps.snapSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(testPublicSnapshot(), nil)
	testDeps.snapSvc.EXPECT().
		GetPrivateSnapshotsByUser(mock.Anything).
		Return(testPrivateSnapshots(), nil)

	testDeps.presence.EXPECT().
		GetConnectedPlayers(mock.Anything).
		Return([]string{testUserID, testUserID2})

	// --- capture message ordering ---
	var (
		mutex         sync.Mutex
		broadcastMsgs []string
		writeMsgs     []string
	)

	testDeps.writer.EXPECT().
		Broadcast(mock.Anything, mock.Anything).
		Run(func(_ ctx.GameContext, msg json.RawMessage) {
			mutex.Lock()
			defer mutex.Unlock()

			broadcastMsgs = append(broadcastMsgs, wsMessageType(t, msg))
		}).
		Return().
		Times(4) // 3 public + 1 move history

	testDeps.writer.EXPECT().
		WriteMessage(mock.Anything, mock.Anything).
		Run(func(_ ctx.GameContext, msg json.RawMessage) {
			mutex.Lock()
			defer mutex.Unlock()

			writeMsgs = append(writeMsgs, wsMessageType(t, msg))
		}).
		Return().
		Times(4) // 2 messages (card + mission) * 2 players

	// Use the TestBus to dispatch synchronously through the registered handler.
	bus := events.NewTestBus()
	pub.Register(bus)
	bus.Emit(gameCtx, event)

	// --- verify ordering ---
	// Broadcasts: first 3 are public state (gameState, boardState, playerState),
	// last 1 is moveHistory.
	require.Len(t, broadcastMsgs, 4)
	assert.Equal(t, "gameState", broadcastMsgs[0])
	assert.Equal(t, "boardState", broadcastMsgs[1])
	assert.Equal(t, "playerState", broadcastMsgs[2])
	assert.Equal(t, "moveHistory", broadcastMsgs[3])

	// WriteMessages: per-player private states (card + mission for each player).
	require.Len(t, writeMsgs, 4)
	// Each player gets cardState then missionState. Since map iteration order is
	// non-deterministic, we verify that we got exactly 2 cardState and 2
	// missionState messages.
	cardCount := 0
	missionCount := 0

	for _, msgType := range writeMsgs {
		switch msgType {
		case "cardState":
			cardCount++
		case "missionState":
			missionCount++
		}
	}

	assert.Equal(t, 2, cardCount, "should have 2 cardState messages (one per player)")
	assert.Equal(t, 2, missionCount, "should have 2 missionState messages (one per player)")
}

// ---------------------------------------------------------------------------
// Test: MoveExecuted — public broadcasts come before private writes, which
// come before move log broadcast. This verifies cross-category ordering.
// ---------------------------------------------------------------------------

func TestHandleMoveExecuted_CrossCategoryOrdering(t *testing.T) {
	t.Parallel()

	testDeps := newDeps(t)
	pub := testDeps.newPublisher()
	gameCtx := testGameContext()
	event := testMoveExecutedEvent()

	testDeps.snapSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(testPublicSnapshot(), nil)
	testDeps.snapSvc.EXPECT().
		GetPrivateSnapshotsByUser(mock.Anything).
		Return(testPrivateSnapshots(), nil)

	testDeps.presence.EXPECT().
		GetConnectedPlayers(mock.Anything).
		Return([]string{testUserID, testUserID2})

	// Track global call sequence across both Broadcast and WriteMessage.
	var (
		mutex    sync.Mutex
		sequence []string // "broadcast:<type>" or "write:<type>"
	)

	testDeps.writer.EXPECT().
		Broadcast(mock.Anything, mock.Anything).
		Run(func(_ ctx.GameContext, msg json.RawMessage) {
			mutex.Lock()
			defer mutex.Unlock()

			sequence = append(sequence, "broadcast:"+wsMessageType(t, msg))
		}).
		Return().
		Times(4)

	testDeps.writer.EXPECT().
		WriteMessage(mock.Anything, mock.Anything).
		Run(func(_ ctx.GameContext, msg json.RawMessage) {
			mutex.Lock()
			defer mutex.Unlock()

			sequence = append(sequence, "write:"+wsMessageType(t, msg))
		}).
		Return().
		Times(4)

	bus := events.NewTestBus()
	pub.Register(bus)
	bus.Emit(gameCtx, event)

	require.Len(t, sequence, 8)

	// First 3 must be public broadcasts.
	assert.Equal(t, "broadcast:gameState", sequence[0])
	assert.Equal(t, "broadcast:boardState", sequence[1])
	assert.Equal(t, "broadcast:playerState", sequence[2])

	// Items 3..6 must be private writes (card/mission per player).
	for i := 3; i < 7; i++ {
		assert.Truef(t, sequence[i] == "write:cardState" || sequence[i] == "write:missionState",
			"index %d should be a private write, got %s", i, sequence[i])
	}

	// Last must be the move history broadcast.
	assert.Equal(t, "broadcast:moveHistory", sequence[7])
}

// ---------------------------------------------------------------------------
// Test: Panic recovery — public state panic does not block private+move log
// ---------------------------------------------------------------------------

func TestHandleMoveExecuted_PanicInPublicState_DoesNotBlockOtherOps(t *testing.T) {
	t.Parallel()

	testDeps := newDeps(t)
	pub := testDeps.newPublisher()
	gameCtx := testGameContext()
	event := testMoveExecutedEvent()

	// GetPublicSnapshot panics — this kills the first safeOp.
	testDeps.snapSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Run(func(_ ctx.GameContext) {
			panic("simulated snapshot service explosion")
		}).
		Return(nil, nil)

	// Private states still runs (second safeOp).
	testDeps.snapSvc.EXPECT().
		GetPrivateSnapshotsByUser(mock.Anything).
		Return(testPrivateSnapshots(), nil)

	// Private state writes per player (card + mission for each user).
	testDeps.writer.EXPECT().
		WriteMessage(mock.Anything, mock.Anything).
		Return().
		Times(4)

	// Move log broadcast still runs (third safeOp) — needs the move log conversion
	// which uses the logging service's ConvertMoveLogs method indirectly through
	// controller.MoveLogController.ConvertMoveLogs.
	// The controller's ConvertMoveLogs does the conversion internally without
	// calling the logging service, so we only need to set up the writer mock.
	testDeps.writer.EXPECT().
		Broadcast(mock.Anything, mock.Anything).
		Return().
		Times(1) // Only the move history broadcast (public state panicked)

	bus := events.NewTestBus()
	pub.Register(bus)

	// Should not panic — safeOp catches it.
	require.NotPanics(t, func() {
		bus.Emit(gameCtx, event)
	})

	// Mock cleanup assertions verify that WriteMessage and Broadcast were called
	// the expected number of times.
}

// ---------------------------------------------------------------------------
// Test: Panic recovery — private state panic does not block move log
// ---------------------------------------------------------------------------

func TestHandleMoveExecuted_PanicInPrivateState_DoesNotBlockMoveLog(t *testing.T) {
	t.Parallel()

	testDeps := newDeps(t)
	pub := testDeps.newPublisher()
	gameCtx := testGameContext()
	event := testMoveExecutedEvent()

	// Public state succeeds.
	testDeps.snapSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(testPublicSnapshot(), nil)
	testDeps.presence.EXPECT().
		GetConnectedPlayers(mock.Anything).
		Return([]string{testUserID})

	// Private state panics.
	testDeps.snapSvc.EXPECT().
		GetPrivateSnapshotsByUser(mock.Anything).
		Run(func(_ ctx.GameContext) {
			panic("simulated private snapshot explosion")
		}).
		Return(nil, nil)

	// Public broadcasts still happen (3 public + 1 move history).
	testDeps.writer.EXPECT().
		Broadcast(mock.Anything, mock.Anything).
		Return().
		Times(4) // 3 public + 1 move log

	bus := events.NewTestBus()
	pub.Register(bus)

	require.NotPanics(t, func() {
		bus.Emit(gameCtx, event)
	})
}

// ---------------------------------------------------------------------------
// Test: GameCompleted calls Lifecycle.RemoveGame
// ---------------------------------------------------------------------------

func TestHandleGameCompleted_CallsRemoveGame(t *testing.T) {
	t.Parallel()

	testDeps := newDeps(t)
	pub := testDeps.newPublisher()
	gameCtx := testGameContext()

	event := gameevt.NewGameCompleted(testGameID, testUserID, time.Now(), 10)

	testDeps.lifecycle.EXPECT().
		RemoveGame(mock.Anything).
		Run(func(gc ctx.GameContext) {
			assert.Equal(t, testGameID, gc.GameID())
		}).
		Return().
		Once()

	bus := events.NewTestBus()
	pub.Register(bus)
	bus.Emit(gameCtx, event)
}

// ---------------------------------------------------------------------------
// Test: PlayerConnected uses WriteMessage (not Broadcast)
// ---------------------------------------------------------------------------

func TestHandlePlayerConnected_UsesWriteMessage(t *testing.T) {
	t.Parallel()

	testDeps := newDeps(t)
	pub := testDeps.newPublisher()
	gameCtx := testGameContext()
	event := gameevt.NewPlayerConnected(testGameID, testUserID, time.Now())

	// Public state — delivered via WriteMessage for the connecting player.
	testDeps.snapSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(testPublicSnapshot(), nil)
	testDeps.presence.EXPECT().
		GetConnectedPlayers(mock.Anything).
		Return([]string{testUserID})

	// Private state for connecting player only.
	testDeps.snapSvc.EXPECT().
		GetPrivateSnapshotsByUser(mock.Anything).
		Return(map[string]*snapshot.PrivateSnapshot{
			testUserID: {
				Cards:       nil,
				MissionType: sqlc.GameMissionTypeTWENTYFOURTERRITORIES,
				MissionID:   1,
			},
		}, nil)

	// Move history for connecting player.
	testDeps.loggingSvc.EXPECT().
		GetMoveLogs(mock.Anything, historySize).
		Return(nil, nil)

	// Capture all WriteMessage calls.
	var (
		mutex    sync.Mutex
		msgTypes []string
	)

	testDeps.writer.EXPECT().
		WriteMessage(mock.Anything, mock.Anything).
		Run(func(_ ctx.GameContext, msg json.RawMessage) {
			mutex.Lock()
			defer mutex.Unlock()

			msgTypes = append(msgTypes, wsMessageType(t, msg))
		}).
		Return().
		Times(6) // 3 public + 2 private + 1 move history

	bus := events.NewTestBus()
	pub.Register(bus)
	bus.Emit(gameCtx, event)

	require.Len(t, msgTypes, 6)

	// First 3: public state via WriteMessage (not Broadcast).
	assert.Equal(t, "gameState", msgTypes[0])
	assert.Equal(t, "boardState", msgTypes[1])
	assert.Equal(t, "playerState", msgTypes[2])

	// 4-5: private state for connecting player.
	assert.Equal(t, "cardState", msgTypes[3])
	assert.Equal(t, "missionState", msgTypes[4])

	// 6: move history.
	assert.Equal(t, "moveHistory", msgTypes[5])
}

// ---------------------------------------------------------------------------
// Test: PlayerConnected ordering — public before private before history
// ---------------------------------------------------------------------------

func TestHandlePlayerConnected_Ordering(t *testing.T) {
	t.Parallel()

	testDeps := newDeps(t)
	pub := testDeps.newPublisher()
	gameCtx := testGameContext()
	event := gameevt.NewPlayerConnected(testGameID, testUserID, time.Now())

	testDeps.snapSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(testPublicSnapshot(), nil)
	testDeps.presence.EXPECT().
		GetConnectedPlayers(mock.Anything).
		Return([]string{testUserID})

	testDeps.snapSvc.EXPECT().
		GetPrivateSnapshotsByUser(mock.Anything).
		Return(map[string]*snapshot.PrivateSnapshot{
			testUserID: {
				Cards:       nil,
				MissionType: sqlc.GameMissionTypeTWENTYFOURTERRITORIES,
				MissionID:   1,
			},
		}, nil)

	testDeps.loggingSvc.EXPECT().
		GetMoveLogs(mock.Anything, historySize).
		Return(nil, nil)

	var (
		mutex    sync.Mutex
		sequence []string
	)

	testDeps.writer.EXPECT().
		WriteMessage(mock.Anything, mock.Anything).
		Run(func(_ ctx.GameContext, msg json.RawMessage) {
			mutex.Lock()
			defer mutex.Unlock()

			sequence = append(sequence, wsMessageType(t, msg))
		}).
		Return().
		Times(6)

	bus := events.NewTestBus()
	pub.Register(bus)
	bus.Emit(gameCtx, event)

	require.Len(t, sequence, 6)

	// Phase 1: public state (gameState, boardState, playerState).
	assert.Equal(t, "gameState", sequence[0])
	assert.Equal(t, "boardState", sequence[1])
	assert.Equal(t, "playerState", sequence[2])

	// Phase 2: private state (cardState, missionState).
	assert.Equal(t, "cardState", sequence[3])
	assert.Equal(t, "missionState", sequence[4])

	// Phase 3: move history.
	assert.Equal(t, "moveHistory", sequence[5])
}

// ---------------------------------------------------------------------------
// Test: PlayerConnected — panic in public state doesn't block private + history
// ---------------------------------------------------------------------------

func TestHandlePlayerConnected_PanicInPublicState_DoesNotBlockOtherOps(t *testing.T) {
	t.Parallel()

	testDeps := newDeps(t)
	pub := testDeps.newPublisher()
	gameCtx := testGameContext()
	event := gameevt.NewPlayerConnected(testGameID, testUserID, time.Now())

	// Public state panics.
	testDeps.snapSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Run(func(_ ctx.GameContext) {
			panic("simulated public snapshot explosion")
		}).
		Return(nil, nil)

	// Private state still runs.
	testDeps.snapSvc.EXPECT().
		GetPrivateSnapshotsByUser(mock.Anything).
		Return(map[string]*snapshot.PrivateSnapshot{
			testUserID: {
				Cards:       nil,
				MissionType: sqlc.GameMissionTypeTWENTYFOURTERRITORIES,
				MissionID:   1,
			},
		}, nil)

	// Move history still runs.
	testDeps.loggingSvc.EXPECT().
		GetMoveLogs(mock.Anything, historySize).
		Return(nil, nil)

	// Private state (2) + move history (1) = 3 WriteMessages.
	testDeps.writer.EXPECT().
		WriteMessage(mock.Anything, mock.Anything).
		Return().
		Times(3)

	bus := events.NewTestBus()
	pub.Register(bus)

	require.NotPanics(t, func() {
		bus.Emit(gameCtx, event)
	})
}

// ---------------------------------------------------------------------------
// Test: GameCompleted — panic in RemoveGame is recovered
// ---------------------------------------------------------------------------

func TestHandleGameCompleted_PanicRecovery(t *testing.T) {
	t.Parallel()

	d := newDeps(t)
	pub := d.newPublisher()
	gameCtx := testGameContext()
	event := gameevt.NewGameCompleted(testGameID, testUserID, time.Now(), 10)

	d.lifecycle.EXPECT().
		RemoveGame(mock.Anything).
		Run(func(_ ctx.GameContext) {
			panic("simulated lifecycle explosion")
		}).
		Return()

	bus := events.NewTestBus()
	pub.Register(bus)

	require.NotPanics(t, func() {
		bus.Emit(gameCtx, event)
	})
}

// ---------------------------------------------------------------------------
// Test: MoveExecuted with real mission resolution
// ---------------------------------------------------------------------------

func TestHandleMoveExecuted_WithMissionResolution(t *testing.T) {
	t.Parallel()

	testDeps := newDeps(t)
	pub := testDeps.newPublisher()
	gameCtx := testGameContext()
	event := testMoveExecutedEvent()

	// Use a mission type that requires service call.
	privateSnaps := map[string]*snapshot.PrivateSnapshot{
		testUserID: {
			Cards:       nil,
			MissionType: sqlc.GameMissionTypeTWOCONTINENTS,
			MissionID:   10,
		},
	}

	testDeps.snapSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(testPublicSnapshot(), nil)
	testDeps.snapSvc.EXPECT().
		GetPrivateSnapshotsByUser(mock.Anything).
		Return(privateSnaps, nil)

	testDeps.presence.EXPECT().
		GetConnectedPlayers(mock.Anything).
		Return([]string{testUserID})

	// Mission service will be called for TWO_CONTINENTS mission.
	testDeps.missionSvc.EXPECT().
		GetTwoContinentsMission(mock.Anything, int64(10)).
		Return(mission.TwoContinentsMission{
			Continent1: "Europe",
			Continent2: "Asia",
		}, nil)

	// 3 public broadcasts + 1 move history = 4 Broadcast calls.
	testDeps.writer.EXPECT().
		Broadcast(mock.Anything, mock.Anything).
		Return().
		Times(4)

	// 2 private writes (card + mission) for one player.
	testDeps.writer.EXPECT().
		WriteMessage(mock.Anything, mock.Anything).
		Return().
		Times(2)

	bus := events.NewTestBus()
	pub.Register(bus)
	bus.Emit(gameCtx, event)
}

// ---------------------------------------------------------------------------
// Test: MoveExecuted — Broadcast is NOT called when Broadcast context is
// wrong, verifying we're not leaking calls between handlers.
// ---------------------------------------------------------------------------

func TestHandlePlayerConnected_NoBroadcastCalls(t *testing.T) {
	t.Parallel()

	testDeps := newDeps(t)
	pub := testDeps.newPublisher()
	gameCtx := testGameContext()
	event := gameevt.NewPlayerConnected(testGameID, testUserID, time.Now())

	testDeps.snapSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(testPublicSnapshot(), nil)
	testDeps.presence.EXPECT().
		GetConnectedPlayers(mock.Anything).
		Return([]string{testUserID})
	testDeps.snapSvc.EXPECT().
		GetPrivateSnapshotsByUser(mock.Anything).
		Return(map[string]*snapshot.PrivateSnapshot{
			testUserID: {
				Cards:       nil,
				MissionType: sqlc.GameMissionTypeTWENTYFOURTERRITORIES,
				MissionID:   1,
			},
		}, nil)
	testDeps.loggingSvc.EXPECT().
		GetMoveLogs(mock.Anything, historySize).
		Return(nil, nil)

	// All messages go through WriteMessage, never Broadcast.
	testDeps.writer.EXPECT().
		WriteMessage(mock.Anything, mock.Anything).
		Return().
		Times(6)

	// Broadcast should NOT be called — if it is, the mock will fail
	// because there is no expectation set for it.

	bus := events.NewTestBus()
	pub.Register(bus)
	bus.Emit(gameCtx, event)
}
