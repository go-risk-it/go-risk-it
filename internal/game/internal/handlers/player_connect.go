package handlers

import (
	"fmt"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/safego"
	"go.uber.org/fx"
)

// playerConnectHandler listens for PlayerConnected events and sends the full game
// state to the connecting player. It reads from the in-memory StateStore first
// (fast path) and falls back to the SnapshotReader (DB) on cache miss.
type playerConnectHandler struct {
	stateStore     gameapi.StateStore
	snapshotReader gameapi.SnapshotReader
	publisher      gameapi.StatePublisher
}

// PlayerConnectHandlerParams holds the dependencies for the player connect handler.
type PlayerConnectHandlerParams struct {
	fx.In

	Sub            eventbus.Subscriber
	StateStore     gameapi.StateStore
	SnapshotReader gameapi.SnapshotReader
	Publisher      gameapi.StatePublisher
}

// RegisterPlayerConnectHandler subscribes the player connect handler to
// PlayerConnected events.
func RegisterPlayerConnectHandler(params PlayerConnectHandlerParams) {
	handler := &playerConnectHandler{
		stateStore:     params.StateStore,
		snapshotReader: params.SnapshotReader,
		publisher:      params.Publisher,
	}

	gameevt.OnGameEvent[*gameevt.PlayerConnected](params.Sub, handler.handlePlayerConnected)
}

func (h *playerConnectHandler) handlePlayerConnected(
	gameCtx gamectx.GameContext,
	_ *gameevt.PlayerConnected,
) {
	safego.TypedSafeOp(gameCtx, "player.connect.state", func(ctx gamectx.GameContext) error {
		playerID := ctx.UserID()
		gameID := ctx.GameID()

		public, private, err := h.resolveState(ctx, gameID, playerID)
		if err != nil {
			return err
		}

		if public == nil || private == nil {
			return nil
		}

		view := snapshot.BuildPlayerView(public, private)

		return h.publisher.PublishState(ctx, playerID, view)
	})
}

// resolveState attempts the StateStore fast path, falling back to the DB reader.
func (h *playerConnectHandler) resolveState(
	ctx gamectx.GameContext,
	gameID int64,
	playerID string,
) (*snapshot.GameSnapshot, *snapshot.PlayerPrivate, error) {
	// Fast path: read from in-memory cache.
	if cached := h.stateStore.Get(gameID); cached != nil {
		if private, ok := cached.PrivateSnapshots[playerID]; ok {
			return cached.PublicSnapshot, private, nil
		}
	}

	// Slow path: read from database.
	public, err := h.snapshotReader.GetPublicSnapshot(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("player connect: get public snapshot: %w", err)
	}

	allPrivate, err := h.snapshotReader.GetAllPrivateSnapshots(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("player connect: get private snapshots: %w", err)
	}

	return public, allPrivate[playerID], nil
}
