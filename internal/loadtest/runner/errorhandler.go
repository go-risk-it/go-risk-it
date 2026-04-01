package runner

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
)

// ErrorHandler implements error recovery with retry logic and escalation.
type ErrorHandler struct {
	gameCtx                 *GameSession
	timeouts                Timeouts
	result                  *GameResult
	ctx                     context.Context
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
	evt := e.(MoveFailedEvent)

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
		h.handleStale(bus, evt)
	case "execution":
		h.handleExecution(bus, evt)
	default:
		// strategy, transient, etc. — just wait and retry.
		h.waitAndEmitState(bus)
	}
}

func (h *ErrorHandler) handleStale(bus *Bus, evt MoveFailedEvent) {
	h.consecutiveStaleErrors++

	if h.consecutiveStaleErrors >= h.maxStaleRetries {
		log.Printf(
			"[game %d] %d stale retries exhausted, advancing",
			h.gameCtx.GameIndex, h.maxStaleRetries,
		)

		phase := h.currentPhase()
		activeREST := h.activeREST()

		if advErr := activeREST.Advance(h.gameCtx.GameID, phase); advErr != nil {
			h.consecutiveAdvanceFails++
			log.Printf(
				"[game %d] advance past %s failed (%d/%d): %v",
				h.gameCtx.GameIndex,
				phase,
				h.consecutiveAdvanceFails,
				h.maxAdvanceFails,
				advErr,
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
		h.waitAndEmitState(bus)

		return
	}

	// Below threshold: wait for update and re-decide.
	_ = evt
	h.waitAndEmitState(bus)
}

func (h *ErrorHandler) handleExecution(bus *Bus, evt MoveFailedEvent) {
	// If card play failed, advance past cards phase.
	if evt.Action != nil && evt.Action.Type == player.ActionPlayCards {
		log.Printf("[game %d] card play failed, advancing past cards phase", h.gameCtx.GameIndex)

		activeREST := h.activeREST()
		if advErr := activeREST.Advance(
			h.gameCtx.GameID,
			string(gamestate.Cards),
		); advErr != nil {
			log.Printf("[game %d] advance past cards also failed: %v", h.gameCtx.GameIndex, advErr)
		} else {
			h.result.Moves++
		}
	}

	h.waitAndEmitState(bus)
}

func (h *ErrorHandler) waitAndEmitState(bus *Bus) {
	waitForAnyUpdate(h.gameCtx.Players, h.timeouts.UpdateWait, h.ctx)
	time.Sleep(h.timeouts.PostMoveSettle)

	snap := h.gameCtx.Players[0].WS.View().Snapshot()
	bus.Emit(StateReceivedEvent{
		Snapshot:  snap,
		Timestamp: time.Now(),
	})
}

func (h *ErrorHandler) currentPhase() string {
	snap := h.gameCtx.Players[0].WS.View().Snapshot()

	return strings.ToLower(string(snap.CurrentPhase()))
}

func (h *ErrorHandler) activeREST() RESTClient {
	// Use first player's REST as fallback.
	return h.gameCtx.Players[0].REST
}
