package request

type ReinforceMove struct {
	SourceRegionID string `json:"sourceRegionId"`
	TargetRegionID string `json:"targetRegionId"`
	TroopsInSource int64  `json:"troopsInSource"`
	TroopsInTarget int64  `json:"troopsInTarget"`
	MovingTroops   int64  `json:"movingTroops"`
}
