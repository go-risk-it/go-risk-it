package request

type DeployMove struct {
	RegionID      string `json:"regionId"`
	CurrentTroops int64  `json:"currentTroops"`
	DesiredTroops int64  `json:"desiredTroops"`
}
