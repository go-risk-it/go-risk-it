package messaging

type PhaseType string

const (
	Cards   PhaseType = "cards"
	Deploy  PhaseType = "deploy"
	Attack  PhaseType = "attack"
	Conquer PhaseType = "conquer"
)

type EmptyState struct{}

type DeployPhaseState struct {
	DeployableTroops int64 `json:"deployableTroops"`
}

type ConquerPhase struct {
	AttackingRegionID string `json:"attackingRegionId"`
	DefendingRegionID string `json:"defendingRegionId"`
	MinTroopsToMove   int64  `json:"minTroopsToMove"`
}

type PhaseState interface {
	EmptyState | DeployPhaseState | ConquerPhase
}

type Phase[T PhaseState] struct {
	Type  PhaseType `json:"type"`
	State T         `json:"state"`
}

type GameState[T PhaseState] struct {
	ID    int64    `json:"id"`
	Turn  int64    `json:"turn"`
	Phase Phase[T] `json:"phase"`
}
