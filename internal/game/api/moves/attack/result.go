// Package attack contains the attack move result DTO used in event payloads.
//
// This type originated in game/logic/move/attack/ and was migrated to game/api/
// to break game/events/ dependencies on the logic layer.
package attack

// MoveResult carries the outcome of an attack move.
type MoveResult struct {
	AttackingRegionID string `json:"attackingRegionId"`
	DefendingRegionID string `json:"defendingRegionId"`
	ConqueringTroops  int64  `json:"conqueringTroops"`
}
