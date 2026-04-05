package weblobby

import (
	"context"
	"fmt"

	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/lobby/internal/logic/start"
	"go.uber.org/fx"
)

// gameCreationConsumer listens for GameCreated and GameCreationFailed events
// from the game module and resolves PendingStarts so the lobby start HTTP
// handler can return the game ID (or an error) to the client.
type gameCreationConsumer struct {
	pending *start.PendingStarts
}

// GameCreationConsumerParams holds the dependencies for the game creation consumer.
type GameCreationConsumerParams struct {
	fx.In

	Sub     eventbus.Subscriber
	Pending *start.PendingStarts
}

// RegisterGameCreationConsumer subscribes the consumer to game creation events.
func RegisterGameCreationConsumer(params GameCreationConsumerParams) {
	c := &gameCreationConsumer{
		pending: params.Pending,
	}

	eventbus.OnEvent(params.Sub, c.handleGameCreated)
	eventbus.OnEvent(params.Sub, c.handleGameCreationFailed)
}

func (c *gameCreationConsumer) handleGameCreated(
	_ context.Context,
	event *gameevt.GameCreated,
) {
	lobbyID := event.LobbyID()
	if lobbyID == 0 {
		// Game created without a lobby (e.g., direct API call). Nothing to resolve.
		return
	}

	c.pending.Resolve(lobbyID, event.GameID(), nil)
}

func (c *gameCreationConsumer) handleGameCreationFailed(
	_ context.Context,
	event *gameevt.GameCreationFailed,
) {
	c.pending.Resolve(event.LobbyID(), 0, fmt.Errorf("game creation failed: %s", event.Reason))
}
