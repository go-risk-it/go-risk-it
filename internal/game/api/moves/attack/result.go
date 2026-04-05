package attack

// MoveResult carries the outcome of an attack move.
type MoveResult struct {
	AttackingRegionID string `json:"attackingRegionId"`
	DefendingRegionID string `json:"defendingRegionId"`
	ConqueringTroops  int64  `json:"conqueringTroops"`
}
