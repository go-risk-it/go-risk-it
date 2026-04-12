package metrics

import (
	"context"
	"sync"
	"sync/atomic"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
)

// gameSummary accumulates per-game counts for recording at completion.
type gameSummary struct {
	moves     atomic.Int64
	attacks   atomic.Int64
	turns     atomic.Int64
	headlines atomic.Int64
}

// gameSummaryRecorder subscribes to game events and records summary histograms
// when each game completes. It uses a mutex-protected map keyed by gameID to
// accumulate per-game counts. The entry is created on GameCreated and deleted
// on MoveCompleted (GameOver=true) after recording all 4 histograms.
//
// All handlers use eventbus.OnEvent (not OnGameEvent) because:
//   - GameCreated is emitted with UserContext, not GameContext (no game ID in ctx yet)
//   - The recorder only needs event.GameID(), never ctx.GameID()
//   - OnGameEvent's GameContext assertion silently drops events with non-game contexts
type gameSummaryRecorder struct {
	mu      sync.Mutex
	games   map[int64]*gameSummary
	metrics *GameMetrics
}

// RegisterGameSummaryRecorder subscribes the recorder to all relevant game events.
func RegisterGameSummaryRecorder(gameMetrics *GameMetrics, sub eventbus.Subscriber) {
	recorder := &gameSummaryRecorder{
		games:   make(map[int64]*gameSummary),
		metrics: gameMetrics,
	}

	eventbus.OnEvent[*gameevt.GameCreated](sub, recorder.handleGameCreated)
	eventbus.OnEvent[*gameevt.MoveCompleted](sub, recorder.handleMoveCompleted)

	sub.OnType(gameevt.TypePlayerEliminated, recorder.handleHeadline)
	sub.OnType(gameevt.TypeContinentCaptured, recorder.handleHeadline)
	sub.OnType(gameevt.TypeContinentLost, recorder.handleHeadline)
}

func (r *gameSummaryRecorder) getSummary(gameID int64) *gameSummary {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.games[gameID]
}

func (r *gameSummaryRecorder) handleGameCreated(
	_ context.Context,
	event *gameevt.GameCreated,
) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.games[event.GameID()] = &gameSummary{}
}

func (r *gameSummaryRecorder) handleMoveCompleted(
	ctx context.Context,
	event *gameevt.MoveCompleted,
) {
	summary := r.getSummary(event.GameID())
	if summary == nil {
		return
	}

	// Count moves and attacks.
	summary.moves.Add(1)

	if event.ActionType == gameapi.GamePhaseTypeATTACK {
		summary.attacks.Add(1)
	}

	// Count turn boundaries (phase transitions from the move's phase).
	if event.FromPhase != event.TargetPhase {
		summary.turns.Add(1)
	}

	// On game completion, record histograms and clean up.
	if event.GameOver {
		r.mu.Lock()
		delete(r.games, event.GameID())
		r.mu.Unlock()

		r.metrics.SummaryMoves.Record(ctx, float64(summary.moves.Load()))
		r.metrics.SummaryAttacks.Record(ctx, float64(summary.attacks.Load()))
		r.metrics.SummaryTurns.Record(ctx, float64(summary.turns.Load()))
		r.metrics.SummaryHeadlines.Record(ctx, float64(summary.headlines.Load()))
	}
}

// handleHeadline handles all headline event types (PlayerEliminated, ContinentCaptured,
// ContinentLost) via a bus.Handler that extracts the gameID from the GameEvent interface.
func (r *gameSummaryRecorder) handleHeadline(_ context.Context, event eventbus.Event) {
	gameEvent, isGameEvent := event.(gameevt.GameEvent)
	if !isGameEvent {
		return
	}

	summary := r.getSummary(gameEvent.GameID())
	if summary == nil {
		return
	}

	summary.headlines.Add(1)
}
