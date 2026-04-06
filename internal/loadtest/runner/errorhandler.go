package runner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
	"go.opentelemetry.io/otel/attribute"
)

// ErrorHandler implements error recovery with retry logic and escalation.
type ErrorHandler struct {
	gameCtx                 *GameSession
	timeouts                Timeouts
	result                  *GameResult
	consecutiveErrors       int
	consecutiveStaleErrors  int
	consecutiveAdvanceFails int
	maxStaleRetries         int
	maxAdvanceFails         int
}

// Register subscribes to EventMoveFailed and EventMoveSucceeded.
func (h *ErrorHandler) Register(bus *Bus) {
	bus.On(EventMoveFailed, h.handleFailed)
	bus.On(EventMoveSucceeded, h.handleSucceeded)
}

func (h *ErrorHandler) handleSucceeded(_ *Bus, _ Event) {
	h.consecutiveErrors = 0
	h.consecutiveStaleErrors = 0
	h.consecutiveAdvanceFails = 0
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

	if h.consecutiveErrors > h.timeouts.MaxConsecutiveErr {
		h.result.FatalError = errors.New("too many consecutive errors")
		h.result.Duration = time.Since(h.gameCtx.StartTime)

		bus.Emit(GameCompleteEvent{Result: *h.result})

		return
	}

	switch evt.ErrType {
	case "stale_state":
		h.handleStale(bus)
	case "execution":
		h.handleExecution(bus, evt)
	default:
		// strategy, transient, etc. — wait for fresh state and retry.
		h.waitFreshAndEmitState(bus)
	}
}

func (h *ErrorHandler) handleStale(bus *Bus) {
	h.consecutiveStaleErrors++

	if h.consecutiveStaleErrors >= h.maxStaleRetries {
		// Wait for a fresh WS update before reading the phase. The view
		// is stale (that's why we got stale_state errors). Reading it now
		// would send an advance for the wrong phase.
		h.waitForFreshState()
		phase := h.currentPhase()

		observe.Warn(h.gameCtx.Ctx, "stale retries exhausted, advancing",
			attribute.Int("gameIndex", h.gameCtx.GameIndex),
			attribute.Int("max_retries", h.maxStaleRetries),
			attribute.String("phase", phase),
		)

		activeREST := h.activeREST()

		if advErr := activeREST.Advance(
			context.Background(), h.gameCtx.GameID, phase,
		); advErr != nil {
			h.consecutiveAdvanceFails++
			observe.Error(h.gameCtx.Ctx, advErr, "advance past phase failed",
				attribute.Int("gameIndex", h.gameCtx.GameIndex),
				attribute.String("phase", phase),
				attribute.Int("attempt", h.consecutiveAdvanceFails),
				attribute.Int("max_attempts", h.maxAdvanceFails),
			)

			if h.consecutiveAdvanceFails >= h.maxAdvanceFails {
				h.result.FatalError = fmt.Errorf(
					"stuck in %s: %d advance attempts failed",
					phase, h.consecutiveAdvanceFails,
				)
				h.result.Duration = time.Since(h.gameCtx.StartTime)

				bus.Emit(GameCompleteEvent{Result: *h.result})

				return
			}
		} else {
			h.result.Moves++
			h.consecutiveAdvanceFails = 0
		}

		h.consecutiveStaleErrors = 0
		h.waitFreshAndEmitState(bus)

		return
	}

	// Below threshold: wait for fresh update and re-decide.
	h.waitFreshAndEmitState(bus)
}

func (h *ErrorHandler) handleExecution(bus *Bus, evt MoveFailedEvent) {
	if evt.Action != nil && evt.Action.Type == player.ActionPlayCards {
		// Wait for fresh state before deciding whether to advance —
		// the server may have already moved past CARDS.
		h.waitForFreshState()
		phase := h.currentPhase()

		if phase == strings.ToLower(string(snapshot.PhaseCards)) {
			observe.Warn(h.gameCtx.Ctx, "card play failed, advancing past cards phase",
				attribute.Int("gameIndex", h.gameCtx.GameIndex),
			)

			activeREST := h.activeREST()
			if advErr := activeREST.Advance(
				context.Background(),
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
	}

	h.waitFreshAndEmitState(bus)
}

// waitForFreshState blocks until a new WS update arrives for all players.
// Does NOT emit StateReceivedEvent — use this for mid-recovery state refresh
// before taking actions like Advance.
func (h *ErrorHandler) waitForFreshState() {
	preVersions := make([]uint64, len(h.gameCtx.Players))
	for i, p := range h.gameCtx.Players {
		preVersions[i] = p.WS.View().Version()
	}

	waitForAllUpdates(h.gameCtx.Players, preVersions, h.timeouts.UpdateWait, h.gameCtx.Ctx)
}

// waitFreshAndEmitState waits for a fresh WS update and emits StateReceivedEvent
// to re-enter the strategy loop.
func (h *ErrorHandler) waitFreshAndEmitState(bus *Bus) {
	preVersions := make([]uint64, len(h.gameCtx.Players))
	for i, p := range h.gameCtx.Players {
		preVersions[i] = p.WS.View().Version()
	}

	waitAndEmitState(bus, h.gameCtx, h.timeouts, preVersions)
}

func (h *ErrorHandler) currentPhase() string {
	snap := h.gameCtx.Players[0].WS.View().Snapshot()

	return strings.ToLower(string(snap.CurrentPhase()))
}

func (h *ErrorHandler) activeREST() RESTClient {
	// Use first player's REST as fallback.
	return h.gameCtx.Players[0].REST
}
