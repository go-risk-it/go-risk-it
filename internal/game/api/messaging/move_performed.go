package messaging

import (
	"time"

	game "github.com/go-risk-it/go-risk-it/internal/game/api"
)

type MovePerformed struct {
	UserID  string         `json:"userId"`
	Phase   game.PhaseType `json:"phase"`
	Move    any            `json:"move"`
	Result  any            `json:"result"`
	Created time.Time      `json:"created"`
}

type MoveHistory struct {
	Moves []MovePerformed `json:"moves"`
}
