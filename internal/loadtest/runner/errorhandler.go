package runner

import (
	"errors"
	"strings"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
	"go.opentelemetry.io/otel/attribute"
)

// ErrorHandler implements non-blocking error recovery. On any error it
// classifies, optionally attempts a recovery action (like Advance), and
// returns. The next WS update from the barrier naturally drives the retry
// cycle — no polling or blocking waits needed.
type ErrorHandler struct {
	gameCtx           *GameSession
	result            *GameResult
	consecutiveErrors int
	maxConsecutiveErr int
}

// Register subscribes to EventMoveFailed and EventMoveSucceeded.
func (h *ErrorHandler) Register(bus *Bus) {
	bus.On(EventMoveFailed, h.handleFailed)
	bus.On(EventMoveSucceeded, h.handleSucceeded)
}

func (h *ErrorHandler) handleSucceeded(_ *Bus, _ Event) {
	h.consecutiveErrors = 0
}

func (h *ErrorHandler) handleFailed(bus *Bus, e Event) {
	evt, ok := e.(MoveFailedEvent)
	if !ok {
		return
	}

	// Fatal errors skip all retry logic.
	if evt.Fatal {
		bus.Emit(GameCompleteEvent{Result: GameResult{
			GameIndex:  h.gameCtx.GameIndex,
			FatalError: evt.Err,
		}})

		return
	}

	h.consecutiveErrors++
	h.result.Errors++

	if h.consecutiveErrors > h.maxConsecutiveErr {
		h.result.FatalError = errors.New("too many consecutive errors")
		h.result.Duration = time.Since(h.gameCtx.StartTime)

		bus.Emit(GameCompleteEvent{Result: *h.result})

		return
	}

	// Execution errors on card play: try to advance past CARDS.
	if evt.ErrType == "execution" && evt.Action != nil &&
		evt.Action.Type == player.ActionPlayCards {
		h.tryAdvancePastCards()
	}

	// For all error types (stale, execution, strategy, transient):
	// return and let the next WS update drive the retry naturally.
	// The barrier guarantees fresh state before the strategy runs again.
}

func (h *ErrorHandler) tryAdvancePastCards() {
	phase := h.currentPhase()
	if phase != strings.ToLower(string(snapshot.PhaseCards)) {
		return
	}

	observe.Warn(h.gameCtx.Ctx, "card play failed, advancing past cards phase",
		attribute.Int("gameIndex", h.gameCtx.GameIndex),
	)

	activeREST := h.activeREST()
	if advErr := activeREST.Advance(
		h.gameCtx.Ctx,
		h.gameCtx.GameID,
		string(snapshot.PhaseCards),
	); advErr != nil {
		observe.Error(h.gameCtx.Ctx, advErr, "advance past cards also failed",
			attribute.Int("gameIndex", h.gameCtx.GameIndex),
		)
	} else {
		h.result.Moves++
	}
}

func (h *ErrorHandler) currentPhase() string {
	snap := h.gameCtx.Players[0].WS.View().Snapshot()

	return strings.ToLower(string(snap.CurrentPhase()))
}

func (h *ErrorHandler) activeREST() RESTClient {
	return h.gameCtx.Players[0].REST
}
