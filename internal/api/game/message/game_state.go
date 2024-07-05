package message

type Phase string

const (
	Cards   Phase = "cards"
	Deploy  Phase = "deploy"
	Attack  Phase = "attack"
	Conquer Phase = "conquer"
)

type DeployPhase struct {
	DeployableTroops int64 `json:"deployableTroops"`
}

type ConquerPhase struct {
	AttackingRegionID string `json:"attackingRegionId"`
	DefendingRegionID string `json:"defendingRegionId"`
	MinTroopsToMove   int64  `json:"minTroopsToMove"`
}

type GameState struct {
	ID           int64 `json:"id"`
	CurrentTurn  int64 `json:"currentTurn"`
	CurrentPhase Phase `json:"currentPhase"`
}
