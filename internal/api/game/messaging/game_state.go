package messaging

import "github.com/go-risk-it/go-risk-it/internal/api/game"

type EmptyState struct{}

type DeployPhaseState struct {
	DeployableTroops int64 `json:"deployableTroops"`
}

type ConquerPhaseState struct {
	AttackingRegionID string `json:"attackingRegionId"`
	DefendingRegionID string `json:"defendingRegionId"`
	MinTroopsToMove   int64  `json:"minTroopsToMove"`
}

type PhaseState interface {
	EmptyState | DeployPhaseState | ConquerPhaseState
}

type Phase[T PhaseState] struct {
	Type  game.PhaseType `json:"type"`
	State T              `json:"state"`
}

type GameState[T PhaseState] struct {
	ID    int64    `json:"id"`
	Turn  int64    `json:"turn"`
	Phase Phase[T] `json:"phase"`
}
