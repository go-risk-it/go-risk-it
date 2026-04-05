// Package cards contains the cards move result DTOs used in event payloads.
//
// These types originated in game/logic/move/cards/ and were migrated to game/api/
// to break game/events/ dependencies on the logic layer.
package cards

// MoveResult carries the outcome of a cards move.
type MoveResult struct {
	ExtraDeployableTroops int64              `json:"extraDeployableTroops"`
	RegionTroopGrants     []RegionTroopGrant `json:"regionTroopGrants"`
}

// RegionTroopGrant records a troop bonus granted to a specific region.
type RegionTroopGrant struct {
	RegionID                int64  `json:"regionId"`
	RegionExternalReference string `json:"regionExternalReference"`
}
