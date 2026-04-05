package handlers_test

import (
	"context"
	"testing"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/handlers"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

// fakePresence returns a fixed list of connected players.
type fakePresence struct {
	players map[int64][]string
}

func (p *fakePresence) ConnectedPlayers(gameID int64) []string {
	return p.players[gameID]
}

// recordingPublisher records PublishState calls for assertions.
type recordingPublisher struct {
	calls []publishCall
}

type publishCall struct {
	playerID string
	view     *snapshot.PlayerView
}

func (p *recordingPublisher) PublishState(
	_ context.Context,
	playerUserID string,
	view *snapshot.PlayerView,
) error {
	p.calls = append(p.calls, publishCall{playerID: playerUserID, view: view})

	return nil
}

func newMoveCompletedForBroadcast(
	gameID int64,
	public *snapshot.GameSnapshot,
	privates map[string]*snapshot.PlayerPrivate,
) *gameevt.MoveCompleted {
	return gameevt.NewMoveCompleted(
		gameID, testAttacker, fixedTime,
		gameapi.GamePhaseTypeDEPLOY,
		1,
		gameapi.GamePhaseTypeDEPLOY,
		gameapi.GamePhaseTypeDEPLOY,
		false,
		public, privates, nil,
	)
}

func TestStateBroadcaster_SendsPlayerViewToEachConnectedPlayer(t *testing.T) {
	t.Parallel()

	bus := newReentrantBus()
	publisher := &recordingPublisher{}
	presence := &fakePresence{players: map[int64][]string{
		testGameID: {"alice", "bob"},
	}}

	handlers.RegisterStateBroadcaster(handlers.StateBroadcasterParams{
		Sub:       bus,
		Publisher: publisher,
		Presence:  presence,
	})

	public := &snapshot.GameSnapshot{
		Game:    snapshot.GameMeta{ID: testGameID, Turn: 1},
		Phase:   snapshot.Phase{Type: snapshot.PhaseDeploy, State: snapshot.EmptyPhaseState{}},
		Regions: []snapshot.RegionState{{ID: "r1", OwnerID: "alice", Troops: 5}},
		Players: []snapshot.PlayerState{{UserID: "alice", Name: "Alice"}},
	}

	privates := map[string]*snapshot.PlayerPrivate{
		"alice": {
			Cards:   []snapshot.CardState{{ID: 1, Type: snapshot.CardInfantry, Region: "r1"}},
			Mission: snapshot.PlayerMission{Type: snapshot.MissionTwentyFourTerritories},
		},
		"bob": {
			Cards:   []snapshot.CardState{},
			Mission: snapshot.PlayerMission{Type: snapshot.MissionEliminatePlayer},
		},
	}

	event := newMoveCompletedForBroadcast(testGameID, public, privates)

	ctx := gamectx.WithGameID(
		kernelctx.WithUserID(kernelctx.WithSpan(context.Background(), noop.Span{}), testAttacker),
		testGameID,
	)
	bus.Emit(ctx, event)

	require.Len(t, publisher.calls, 2)

	playerIDs := make(map[string]bool)
	for _, call := range publisher.calls {
		playerIDs[call.playerID] = true
		assert.Equal(t, public.Game, call.view.Game)
		assert.Equal(t, public.Phase, call.view.Phase)
	}

	assert.True(t, playerIDs["alice"])
	assert.True(t, playerIDs["bob"])
}

func TestStateBroadcaster_SkipsMissingPrivateSnapshot(t *testing.T) {
	t.Parallel()

	bus := newReentrantBus()
	publisher := &recordingPublisher{}
	presence := &fakePresence{players: map[int64][]string{
		testGameID: {"alice", "unknown"},
	}}

	handlers.RegisterStateBroadcaster(handlers.StateBroadcasterParams{
		Sub:       bus,
		Publisher: publisher,
		Presence:  presence,
	})

	public := &snapshot.GameSnapshot{
		Game: snapshot.GameMeta{ID: testGameID, Turn: 1},
	}

	privates := map[string]*snapshot.PlayerPrivate{
		"alice": {Cards: []snapshot.CardState{}, Mission: snapshot.PlayerMission{}},
		// "unknown" not in privates — should be skipped
	}

	event := newMoveCompletedForBroadcast(testGameID, public, privates)
	bus.Emit(gameCtx(testGameID), event)

	// Only alice gets a message
	require.Len(t, publisher.calls, 1)
	assert.Equal(t, "alice", publisher.calls[0].playerID)
}

func TestStateBroadcaster_NoPlayersConnected(t *testing.T) {
	t.Parallel()

	bus := newReentrantBus()
	publisher := &recordingPublisher{}
	presence := &fakePresence{players: map[int64][]string{}}

	handlers.RegisterStateBroadcaster(handlers.StateBroadcasterParams{
		Sub:       bus,
		Publisher: publisher,
		Presence:  presence,
	})

	event := newMoveCompletedForBroadcast(testGameID, &snapshot.GameSnapshot{}, nil)
	bus.Emit(gameCtx(testGameID), event)

	require.Empty(t, publisher.calls)
}

func TestStateBroadcaster_PlayerViewHasCorrectPrivateData(t *testing.T) {
	t.Parallel()

	bus := newReentrantBus()
	publisher := &recordingPublisher{}
	presence := &fakePresence{players: map[int64][]string{
		testGameID: {"alice"},
	}}

	handlers.RegisterStateBroadcaster(handlers.StateBroadcasterParams{
		Sub:       bus,
		Publisher: publisher,
		Presence:  presence,
	})

	expectedCards := []snapshot.CardState{{ID: 42, Type: snapshot.CardCavalry, Region: "brazil"}}
	expectedMission := snapshot.PlayerMission{
		Type:   snapshot.MissionTwoContinents,
		Detail: snapshot.TwoContinentsMission{Continent1: "europe", Continent2: "asia"},
	}

	public := &snapshot.GameSnapshot{
		Game: snapshot.GameMeta{ID: testGameID, Turn: 3},
	}

	privates := map[string]*snapshot.PlayerPrivate{
		"alice": {Cards: expectedCards, Mission: expectedMission},
	}

	event := newMoveCompletedForBroadcast(testGameID, public, privates)
	bus.Emit(gameCtx(testGameID), event)

	require.Len(t, publisher.calls, 1)
	assert.Equal(t, expectedCards, publisher.calls[0].view.Cards)
	assert.Equal(t, expectedMission, publisher.calls[0].view.Mission)
}

// Ensure the Presence interface is satisfied by fakePresence.
var _ handlers.Presence = (*fakePresence)(nil)

// Ensure the StatePublisher interface is satisfied by recordingPublisher.
var _ gameapi.StatePublisher = (*recordingPublisher)(nil)
