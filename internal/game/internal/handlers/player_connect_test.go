package handlers_test

import (
	"context"
	"errors"
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

// stubSnapshotReader returns fixed public/private snapshots.
type stubSnapshotReader struct {
	public  *snapshot.GameSnapshot
	private map[string]*snapshot.PlayerPrivate
	err     error
}

func (s *stubSnapshotReader) GetPublicSnapshot(
	_ gamectx.GameContext,
) (*snapshot.GameSnapshot, error) {
	return s.public, s.err
}

func (s *stubSnapshotReader) GetAllPrivateSnapshots(
	_ gamectx.GameContext,
) (map[string]*snapshot.PlayerPrivate, error) {
	return s.private, s.err
}

var _ gameapi.SnapshotReader = (*stubSnapshotReader)(nil)

func playerConnectedCtx(gameID int64, userID string) gamectx.GameContext {
	return gamectx.WithGameID(
		kernelctx.WithUserID(kernelctx.WithSpan(context.Background(), noop.Span{}), userID),
		gameID,
	)
}

func TestPlayerConnect_CacheHit(t *testing.T) {
	t.Parallel()

	bus := newReentrantBus()
	publisher := &recordingPublisher{}
	store := handlers.NewStateStore()

	public := &snapshot.GameSnapshot{
		Game: snapshot.GameMeta{ID: testGameID, Turn: 3},
	}
	privates := map[string]*snapshot.PlayerPrivate{
		"alice": {Cards: []snapshot.CardState{{ID: 1}}, Mission: snapshot.PlayerMission{}},
	}

	store.Store(testGameID, &snapshot.CachedGameState{
		Turn:             3,
		PublicSnapshot:   public,
		PrivateSnapshots: privates,
	})

	reader := &stubSnapshotReader{} // should NOT be called

	handlers.RegisterPlayerConnectHandler(handlers.PlayerConnectHandlerParams{
		Sub:            bus,
		StateStore:     store,
		SnapshotReader: reader,
		Publisher:      publisher,
	})

	event := gameevt.NewPlayerConnected(testGameID, "alice", fixedTime)
	bus.Emit(playerConnectedCtx(testGameID, "alice"), event)

	require.Len(t, publisher.calls, 1)
	assert.Equal(t, "alice", publisher.calls[0].playerID)
	assert.Equal(t, public.Game, publisher.calls[0].view.Game)
}

func TestPlayerConnect_CacheMiss_FallsBackToDB(t *testing.T) {
	t.Parallel()

	bus := newReentrantBus()
	publisher := &recordingPublisher{}
	store := handlers.NewStateStore()

	// No cached state — empty store

	public := &snapshot.GameSnapshot{
		Game: snapshot.GameMeta{ID: testGameID, Turn: 1},
	}
	privates := map[string]*snapshot.PlayerPrivate{
		"bob": {Cards: []snapshot.CardState{}, Mission: snapshot.PlayerMission{}},
	}

	reader := &stubSnapshotReader{
		public:  public,
		private: privates,
	}

	handlers.RegisterPlayerConnectHandler(handlers.PlayerConnectHandlerParams{
		Sub:            bus,
		StateStore:     store,
		SnapshotReader: reader,
		Publisher:      publisher,
	})

	event := gameevt.NewPlayerConnected(testGameID, "bob", fixedTime)
	bus.Emit(playerConnectedCtx(testGameID, "bob"), event)

	require.Len(t, publisher.calls, 1)
	assert.Equal(t, "bob", publisher.calls[0].playerID)
}

func TestPlayerConnect_CacheMiss_DBError(t *testing.T) {
	t.Parallel()

	bus := newReentrantBus()
	publisher := &recordingPublisher{}
	store := handlers.NewStateStore()

	reader := &stubSnapshotReader{err: errors.New("db down")}

	handlers.RegisterPlayerConnectHandler(handlers.PlayerConnectHandlerParams{
		Sub:            bus,
		StateStore:     store,
		SnapshotReader: reader,
		Publisher:      publisher,
	})

	event := gameevt.NewPlayerConnected(testGameID, "alice", fixedTime)

	require.NotPanics(t, func() {
		bus.Emit(playerConnectedCtx(testGameID, "alice"), event)
	})

	require.Empty(t, publisher.calls)
}

func TestPlayerConnect_CacheHit_PlayerNotInPrivate(t *testing.T) {
	t.Parallel()

	bus := newReentrantBus()
	publisher := &recordingPublisher{}
	store := handlers.NewStateStore()

	// Cached state exists but connecting player isn't in private snapshots
	store.Store(testGameID, &snapshot.CachedGameState{
		Turn:           1,
		PublicSnapshot: &snapshot.GameSnapshot{},
		PrivateSnapshots: map[string]*snapshot.PlayerPrivate{
			"other": {Cards: []snapshot.CardState{}, Mission: snapshot.PlayerMission{}},
		},
	})

	public := &snapshot.GameSnapshot{
		Game: snapshot.GameMeta{ID: testGameID, Turn: 1},
	}
	privates := map[string]*snapshot.PlayerPrivate{
		"alice": {Cards: []snapshot.CardState{}, Mission: snapshot.PlayerMission{}},
	}

	reader := &stubSnapshotReader{
		public:  public,
		private: privates,
	}

	handlers.RegisterPlayerConnectHandler(handlers.PlayerConnectHandlerParams{
		Sub:            bus,
		StateStore:     store,
		SnapshotReader: reader,
		Publisher:      publisher,
	})

	event := gameevt.NewPlayerConnected(testGameID, "alice", fixedTime)
	bus.Emit(playerConnectedCtx(testGameID, "alice"), event)

	// Falls back to DB, sends view for alice
	require.Len(t, publisher.calls, 1)
	assert.Equal(t, "alice", publisher.calls[0].playerID)
}
