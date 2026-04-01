package runner

import (
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/orchestrator"
)

// HealthHandler delegates game lifecycle events to a GameObserver.
type HealthHandler struct {
	observer  orchestrator.GameObserver
	gameCtx   *GameSession
	lastPhase metrics.Phase
}

// NewHealthHandler creates a HealthHandler. Nil observer uses NopObserver.
func NewHealthHandler(obs orchestrator.GameObserver, gameCtx *GameSession) *HealthHandler {
	if obs == nil {
		obs = orchestrator.NopObserver{}
	}

	return &HealthHandler{observer: obs, gameCtx: gameCtx}
}

// Register subscribes to lifecycle events.
func (h *HealthHandler) Register(bus *Bus) {
	bus.On(EventGameStarted, h.handleStarted)
	bus.On(EventMoveDecided, h.handleMoveDecided)
	bus.On(EventMoveSucceeded, h.handleMoveSucceeded)
	bus.On(EventGameComplete, h.handleComplete)
}

func (h *HealthHandler) handleStarted(_ *Bus, e Event) {
	evt := e.(GameStartedEvent)
	h.observer.OnGameStarted(evt.GameIndex)
}

func (h *HealthHandler) handleMoveDecided(_ *Bus, e Event) {
	evt := e.(MoveDecidedEvent)

	if evt.Phase != h.lastPhase {
		h.observer.OnPhaseChange(h.gameCtx.GameIndex, string(evt.Phase))
		h.lastPhase = evt.Phase
	}
}

func (h *HealthHandler) handleMoveSucceeded(_ *Bus, e Event) {
	evt := e.(MoveSucceededEvent)
	phase := actionTypeName(evt.Action.Type)
	h.observer.OnMove(h.gameCtx.GameIndex, phase)
}

func (h *HealthHandler) handleComplete(_ *Bus, e Event) {
	evt := e.(GameCompleteEvent)
	h.observer.OnGameComplete(evt.Result.GameIndex)
}
