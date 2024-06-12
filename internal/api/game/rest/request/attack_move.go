package request

type AttackMove struct {
	SourceRegionID  string `json:"sourceRegionId"`
	TargetRegionID  string `json:"targetRegionId"`
	TroopsInSource  int64  `json:"troopsInSource"`
	TroopsInTarget  int64  `json:"troopsInTarget"`
	AttackingTroops int64  `json:"attackingTroops"`
}
