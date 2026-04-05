package runner

import (
	"context"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
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

// waitSettleAndEmitState waits for all players' WS views to update, snapshots
// player 0, and emits a StateReceivedEvent.
//
// We wait for ALL players (not just one) because the unified playerView protocol
// sends exactly one WS message per player per move. Waiting on "any" single player
// could return before the others have processed their messages, yielding a stale
// snapshot when reading a different player's view later in the pipeline.
func waitSettleAndEmitState(bus *Bus, gameCtx *GameSession, timeouts Timeouts) {
	waitForAllUpdates(gameCtx.Players, timeouts.UpdateWait, gameCtx.Ctx)

	wsReceivedAt := time.Now()
	time.Sleep(timeouts.PostMoveSettle)

	snap := gameCtx.Players[0].WS.View().Snapshot()
	bus.Emit(StateReceivedEvent{
		Snapshot:     snap,
		Timestamp:    time.Now(),
		WSReceivedAt: wsReceivedAt,
	})
}

// waitForAllUpdates waits until every player's WS connection has received at
// least one update, or the timeout/context expires.
func waitForAllUpdates(players []*PlayerInfo, timeout time.Duration, ctx context.Context) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for _, p := range players {
		select {
		case <-p.WS.View().Updated():
		case <-p.WS.Done():
		case <-timer.C:
			return
		case <-ctx.Done():
			return
		}
	}
}

// waitForPhaseChange waits until the view shows a different phase than oldPhase.
func waitForPhaseChange(
	v *gamestate.View,
	oldPhase snapshot.PhaseType,
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
