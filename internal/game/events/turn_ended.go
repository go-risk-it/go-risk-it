package game

import (
	"time"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
)

// RegisterTurnEndedEmitter subscribes to MoveCompleted events and emits
// TurnEnded when a player's REINFORCE phase ends. REINFORCE is always the last
// phase of a player's turn, so FromPhase==REINFORCE reliably detects turn boundaries
// regardless of whether the transition came from the orchestrator (explicit move)
// or the advancement service (auto-advance).
//
// The TurnEnded timestamp is offset by 1ms from the triggering MoveCompleted.
// This offset is load-bearing: the event logger uses EventTimestamp() as the OTLP
// log record timestamp (not time.Now()), so the 1ms guarantees TurnEnded sorts
// AFTER the preceding REINFORCE MoveCompleted in Loki. Without it, both events
// share the same timestamp and State Timeline forward-fills REINFORCE through
// the idle gap instead of showing WAITING.
func RegisterTurnEndedEmitter(sub eventbus.Subscriber, pub eventbus.Publisher) {
	OnGameEvent[*MoveCompleted](sub,
		func(ctx gamectx.GameContext, event *MoveCompleted) {
			if event.FromPhase == gameapi.GamePhaseTypeREINFORCE {
				pub.Emit(ctx, NewTurnEnded(
					event.GameID(),
					ctx.UserID(),
					event.EventTimestamp().Add(1*time.Millisecond),
					event.Turn,
				))
			}
		},
	)
}
