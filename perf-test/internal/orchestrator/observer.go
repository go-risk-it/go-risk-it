package orchestrator

// GameObserver receives lifecycle events from game execution.
// Implementations must be safe for concurrent use from multiple goroutines.
type GameObserver interface {
	OnGameStarted(gameIndex int)
	OnMove(gameIndex int, phase string)
	OnPhaseChange(gameIndex int, newPhase string)
	OnGameComplete(gameIndex int)
}

// NopObserver is a no-op implementation for backward compatibility.
type NopObserver struct{}

func (NopObserver) OnGameStarted(int)         {}
func (NopObserver) OnMove(int, string)        {}
func (NopObserver) OnPhaseChange(int, string) {}
func (NopObserver) OnGameComplete(int)        {}
