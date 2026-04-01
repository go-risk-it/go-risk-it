package runner //nolint:testpackage // whitebox tests access unexported helpers

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
	"github.com/stretchr/testify/assert"
)

func TestChaos_MoveSucceeded_MaybeDisconnects(t *testing.T) {
	t.Parallel()

	var called bool
	var numPlayers int

	gameCtx := &GameSession{
		Players: []*PlayerInfo{
			{UserID: "u0", Name: "p0", WS: newFakeWS()},
			{UserID: "u1", Name: "p1", WS: newFakeWS()},
			{UserID: "u2", Name: "p2", WS: newFakeWS()},
			{UserID: "u3", Name: "p3", WS: newFakeWS()},
		},
		UserIndex: map[string]int{"u0": 0, "u1": 1, "u2": 2, "u3": 3},
	}

	h := &ChaosHandler{
		maybeDisconnectFn: func(players []*PlayerInfo, _ int) {
			called = true
			numPlayers = len(players)
		},
		gameCtx: gameCtx,
	}
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(MoveSucceededEvent{
		Action: &player.Action{Type: player.ActionDeploy},
	})

	assert.True(t, called)
	assert.Equal(t, 4, numPlayers)
}

func TestChaos_NilInjector_NoRegistration(t *testing.T) {
	t.Parallel()

	gameCtx := &GameSession{
		Players: []*PlayerInfo{
			{UserID: "u0", Name: "p0", WS: newFakeWS()},
		},
	}

	h := NewChaosHandler(nil, gameCtx)
	bus := NewTestBus()
	h.Register(bus)

	// Should not panic.
	bus.Emit(MoveSucceededEvent{
		Action: &player.Action{Type: player.ActionDeploy},
	})
}
