package handlers

import (
	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/kernel/safego"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/fx"
)

// Presence exposes the set of currently connected player user IDs for a game.
type Presence interface {
	ConnectedPlayers(gameID int64) []string
}

// stateBroadcaster listens for MoveCompleted events and sends a per-player
// PlayerView to every connected client via the StatePublisher port.
type stateBroadcaster struct {
	publisher gameapi.StatePublisher
	presence  Presence
}

// StateBroadcasterParams holds the dependencies for the state broadcaster handler.
type StateBroadcasterParams struct {
	fx.In

	Sub       eventbus.Subscriber
	Publisher gameapi.StatePublisher
	Presence  Presence
}

// RegisterStateBroadcaster subscribes the state broadcaster to MoveCompleted events.
func RegisterStateBroadcaster(params StateBroadcasterParams) {
	b := &stateBroadcaster{
		publisher: params.Publisher,
		presence:  params.Presence,
	}

	gameevt.OnGameEvent[*gameevt.MoveCompleted](params.Sub, b.handleMoveCompleted)
}

func (b *stateBroadcaster) handleMoveCompleted(
	gameCtx gamectx.GameContext,
	event *gameevt.MoveCompleted,
) {
	safego.TypedSafeOp(gameCtx, "broadcaster.state", func(ctx gamectx.GameContext) error {
		players := b.presence.ConnectedPlayers(event.GameID())

		for _, playerID := range players {
			private := event.PrivateSnapshots[playerID]
			if private == nil {
				observe.Warn(ctx, "state broadcaster: no private snapshot for player",
					attribute.String("player", playerID))

				continue
			}

			view := snapshot.BuildPlayerView(event.PublicSnapshot, private)
			if err := b.publisher.PublishState(ctx, playerID, view); err != nil {
				observe.Warn(ctx, "state broadcaster: publish failed",
					attribute.String("player", playerID))
			}
		}

		return nil
	})
}
