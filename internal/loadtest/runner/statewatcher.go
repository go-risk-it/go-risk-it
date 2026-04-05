package runner

import (
	"context"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
)

// StateWatcherHandler waits for fresh game state via WebSocket after moves.
// It captures pre-move view versions on MoveDecided (before the REST call),
// then uses them after MoveSucceeded to detect updates that may have arrived
// during the REST roundtrip.
type StateWatcherHandler struct {
	gameCtx  *GameSession
	timeouts Timeouts
}

// Register subscribes to the events that drive state watching.
func (h *StateWatcherHandler) Register(bus *Bus) {
	bus.On(EventMoveSucceeded, h.handleWaitAndEmit)
	bus.On(EventTurnSkipped, h.handleWaitAndEmit)
	bus.On(EventMoveConflict, h.handleConflict)
}

func (h *StateWatcherHandler) handleWaitAndEmit(bus *Bus, _ Event) {
	if h.gameCtx.Ctx.Err() != nil {
		return
	}

	// Capture versions NOW — after the REST call succeeded but before
	// waiting for the WS broadcast. The server has committed the move
	// and will broadcast imminently; any version bump from here onward
	// is the update we're waiting for.
	preVersions := make([]uint64, len(h.gameCtx.Players))
	for i, p := range h.gameCtx.Players {
		preVersions[i] = p.WS.View().Version()
	}

	waitAndEmitState(bus, h.gameCtx, h.timeouts, preVersions)
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

// waitAndEmitState waits for all players' WS views to update past
// their pre-move versions, snapshots player 0, and emits a StateReceivedEvent.
//
// We wait for ALL players (not just one) because the unified playerView protocol
// sends exactly one WS message per player per move. Waiting on "any" single player
// could return before the others have processed their messages, yielding a stale
// snapshot when reading a different player's view later in the pipeline.
func waitAndEmitState(
	bus *Bus,
	gameCtx *GameSession,
	timeouts Timeouts,
	preVersions []uint64,
) {
	waitForAllUpdates(gameCtx.Players, preVersions, timeouts.UpdateWait, gameCtx.Ctx)

	now := time.Now()
	snap := gameCtx.Players[0].WS.View().Snapshot()
	bus.Emit(StateReceivedEvent{
		Snapshot:     snap,
		Timestamp:    now,
		WSReceivedAt: now,
	})
}

// waitForAllUpdates waits until every player's WS connection has received an
// update after the pre-move version snapshot, or the timeout/context expires.
// Using version-based detection eliminates the race where a WS message arrives
// between the REST response and the start of the wait.
func waitForAllUpdates(
	players []*PlayerInfo,
	preVersions []uint64,
	timeout time.Duration,
	ctx context.Context,
) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for i, p := range players {
		var version uint64
		if i < len(preVersions) {
			version = preVersions[i]
		}

		select {
		case <-p.WS.View().AwaitUpdateSince(version):
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
		currentVersion := v.Version()

		if v.Snapshot().CurrentPhase() != oldPhase {
			return
		}

		select {
		case <-deadline:
			return
		case <-v.AwaitUpdateSince(currentVersion):
		case <-ctx.Done():
			return
		}
	}
}
