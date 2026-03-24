package runner

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player"
	"github.com/stretchr/testify/assert"
)

// fakeObserver tracks GameObserver calls.
type fakeObserver struct {
	started      []int
	moves        []string
	phaseChanges []string
	completed    []int
}

func (f *fakeObserver) OnGameStarted(idx int)        { f.started = append(f.started, idx) }
func (f *fakeObserver) OnMove(idx int, phase string) { f.moves = append(f.moves, phase) }
func (f *fakeObserver) OnPhaseChange(idx int, phase string) {
	f.phaseChanges = append(f.phaseChanges, phase)
}
func (f *fakeObserver) OnGameComplete(idx int) { f.completed = append(f.completed, idx) }

func TestHealth_DelegatesAllHooks(t *testing.T) {
	t.Parallel()

	obs := &fakeObserver{}
	gameCtx := &GameContext{GameIndex: 7}
	h := &HealthHandler{observer: obs, gameCtx: gameCtx}
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(GameStartedEvent{GameIndex: 7, NumPlayers: 4})

	// Emit MoveDecided with new phase to trigger phase change.
	bus.Emit(MoveDecidedEvent{
		Action: &player.Action{Type: player.ActionDeploy},
		UserID: "u0",
		Phase:  "deploy",
	})

	bus.Emit(MoveSucceededEvent{
		Action: &player.Action{Type: player.ActionDeploy},
	})

	bus.Emit(GameCompleteEvent{Result: GameResult{GameIndex: 7}})

	assert.Equal(t, []int{7}, obs.started)
	assert.Equal(t, []string{"deploy"}, obs.phaseChanges)
	assert.Len(t, obs.moves, 1)
	assert.Equal(t, []int{7}, obs.completed)
}

func TestHealth_NilObserver_UsesNop(t *testing.T) {
	t.Parallel()

	gameCtx := &GameContext{GameIndex: 1}
	h := NewHealthHandler(nil, gameCtx)
	bus := NewTestBus()
	h.Register(bus)

	// Should not panic.
	bus.Emit(GameStartedEvent{GameIndex: 1, NumPlayers: 4})
	bus.Emit(GameCompleteEvent{Result: GameResult{GameIndex: 1}})

	// Verify it uses NopObserver.
	_, isNop := h.observer.(orchestrator.NopObserver)
	assert.True(t, isNop)
}
