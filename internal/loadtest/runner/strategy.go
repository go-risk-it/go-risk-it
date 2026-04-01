package runner

import (
	"context"
	"strings"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
	"go.opentelemetry.io/otel/attribute"
)

// StrategyHandler receives game state, decides what to do next.
// This is the routing point with 3 possible outputs.
type StrategyHandler struct {
	strategy  player.Strategy
	thinkTime time.Duration
	gameCtx   *GameSession
}

// Register subscribes to EventStateReceived.
func (h *StrategyHandler) Register(bus *Bus) {
	bus.On(EventStateReceived, h.handle)
}

//nolint:funlen // sequential strategy dispatch
func (h *StrategyHandler) handle(bus *Bus, e Event) {
	evt := e.(StateReceivedEvent) //nolint:forcetypeassert // event bus guarantees type
	snap := evt.Snapshot

	// Check context cancellation.
	if h.gameCtx.Ctx.Err() != nil {
		bus.Emit(GameCompleteEvent{Result: GameResult{
			GameIndex: h.gameCtx.GameIndex,
			TimedOut:  h.gameCtx.Ctx.Err() == context.DeadlineExceeded,
		}})

		return
	}

	// Check game over.
	if snap.IsGameOver() {
		bus.Emit(GameCompleteEvent{Result: GameResult{
			GameIndex: h.gameCtx.GameIndex,
			Winner:    snap.GameState.WinnerUserID,
		}})

		return
	}

	// Check nil state.
	if snap.GameState == nil || snap.PlayersState == nil {
		bus.Emit(TurnSkippedEvent{})

		return
	}

	// Find active player.
	activeIdx, activeUserID := findActivePlayer(snap, h.gameCtx.UserIndex)
	if activeIdx < 0 {
		bus.Emit(TurnSkippedEvent{})

		return
	}

	// Use active player's own view for the most accurate state.
	activeSnap := h.gameCtx.Players[activeIdx].WS.View().Snapshot()

	currentPhase := strings.ToLower(string(activeSnap.CurrentPhase()))

	// Decide move.
	action, err := h.strategy.DecideMove(activeSnap, activeUserID)
	if err != nil {
		observe.Error(h.gameCtx.Ctx, err, "strategy error",
			attribute.Int("gameIndex", h.gameCtx.GameIndex),
			attribute.String("player", h.gameCtx.Players[activeIdx].Name),
			attribute.String("phase", string(activeSnap.CurrentPhase())),
		)
		bus.Emit(MoveFailedEvent{
			Err:     err,
			Fatal:   false,
			ErrType: metrics.ErrorTypeStrategy,
		})

		return
	}

	// Apply think time.
	if h.thinkTime > 0 {
		time.Sleep(h.thinkTime)
	}

	bus.Emit(MoveDecidedEvent{
		Action: action,
		UserID: activeUserID,
		Phase:  metrics.Phase(currentPhase),
	})
}

// findActivePlayer finds the player whose turn it is.
// Returns the player index and userID, or (-1, "") if not found.
func findActivePlayer(
	snap gamestate.ViewSnapshot,
	userIndex map[string]int,
) (int, string) {
	if snap.GameState == nil || snap.PlayersState == nil {
		return -1, ""
	}

	numPlayers := int64(len(snap.PlayersState.Players))
	if numPlayers == 0 {
		return -1, ""
	}

	currentIndex := snap.GameState.Turn % numPlayers
	for _, p := range snap.PlayersState.Players {
		if p.Index == currentIndex {
			if idx, ok := userIndex[p.UserID]; ok {
				return idx, p.UserID
			}
		}
	}

	return -1, ""
}

// actionTypeName returns the metric name for an action type.
func actionTypeName(t player.ActionType) string {
	switch t {
	case player.ActionDeploy:
		return "deploy"
	case player.ActionAttack:
		return "attack"
	case player.ActionConquer:
		return "conquer"
	case player.ActionReinforce:
		return "reinforce"
	case player.ActionPlayCards:
		return "cards"
	case player.ActionAdvance:
		return "advance"
	default:
		return "unknown"
	}
}
