package signals

import (
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/maniartech/signals"
)

type GameStateChangedData struct {
	FromPhase sqlc.GamePhaseType
	ToPhase   sqlc.GamePhaseType
}

type GameStateChangedSignal interface {
	signals.Signal[GameStateChangedData]
}

func NewGameStateChangedSignal() GameStateChangedSignal {
	return signals.New[GameStateChangedData]()
}
