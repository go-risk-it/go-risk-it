package runner

import (
	"context"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
)

// StateWatcherHandler waits for fresh game state via WebSocket after moves.
type StateWatcherHandler struct {
	gameCtx  *GameSession
	timeouts Timeouts
}

// Register subscribes to EventMoveSucceeded, EventMoveConflict, EventTurnSkipped.
func (h *StateWatcherHandler) Register(bus *Bus) {
	bus.On(EventMoveSucceeded, h.handleWaitAndEmit)
	bus.On(EventTurnSkipped, h.handleWaitAndEmit)
	bus.On(EventMoveConflict, h.handleConflict)
}

func (h *StateWatcherHandler) handleWaitAndEmit(bus *Bus, _ Event) {
	if h.gameCtx.Ctx.Err() != nil {
		return
	}

	waitSettleAndEmitState(bus, h.gameCtx, h.timeouts)
}

func (h *StateWatcherHandler) handleConflict(bus *Bus, _ Event) {
	if h.gameCtx.Ctx.Err() != nil {
		return
	}

	// Wait for phase to change (indicates server moved past the conflict).
	activeView := h.gameCtx.Players[0].WS.View()
	oldPhase := activeView.Snapshot().CurrentPhase()
	waitForPhaseChange(activeView, oldPhase, h.timeouts.PhaseChangeWait, h.gameCtx.Ctx)

	if h.gameCtx.Ctx.Err() != nil {
		return
	}

	snap := activeView.Snapshot()
	bus.Emit(StateReceivedEvent{
		Snapshot:  snap,
		Timestamp: time.Now(),
	})
}

// waitSettleAndEmitState waits for a WS state update, sleeps the settle time,
// snapshots player 0, and emits a StateReceivedEvent. Shared by ErrorHandler
// and StateWatcherHandler.
func waitSettleAndEmitState(bus *Bus, gameCtx *GameSession, timeouts Timeouts) {
	waitForAnyUpdate(gameCtx.Players, timeouts.UpdateWait, gameCtx.Ctx)
	time.Sleep(timeouts.PostMoveSettle)

	snap := gameCtx.Players[0].WS.View().Snapshot()
	bus.Emit(StateReceivedEvent{
		Snapshot:  snap,
		Timestamp: time.Now(),
	})
}

// waitForAnyUpdate waits for a state update from any player's WS connection.
//
//nolint:cyclop // multi-player state coordination
func waitForAnyUpdate(players []*PlayerInfo, timeout time.Duration, ctx context.Context) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	done := make(chan struct{})
	defer close(done)

	signal := make(chan struct{}, 1)
	for _, p := range players {
		go func(updated <-chan struct{}, wsDone <-chan struct{}) {
			select {
			case <-updated:
				select {
				case signal <- struct{}{}:
				default:
				}
			case <-wsDone:
				select {
				case signal <- struct{}{}:
				default:
				}
			case <-done:
			}
		}(p.WS.View().Updated(), p.WS.Done())
	}

	select {
	case <-signal:
	case <-timer.C:
	case <-ctx.Done():
	}
}

// waitForPhaseChange waits until the view shows a different phase than oldPhase.
func waitForPhaseChange(
	v *gamestate.View,
	oldPhase gamestate.PhaseType,
	timeout time.Duration,
	ctx context.Context,
) {
	deadline := time.After(timeout)

	for {
		select {
		case <-deadline:
			return
		case <-v.Updated():
			if v.Snapshot().CurrentPhase() != oldPhase {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
