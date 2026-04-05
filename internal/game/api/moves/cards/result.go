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
