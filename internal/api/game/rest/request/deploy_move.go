package request

type DeployMove struct {
	UserID        string `json:"userId"`
	RegionID      string `json:"regionId"`
	CurrentTroops int64  `json:"currentTroops"`
	DesiredTroops int64  `json:"desiredTroops"`
}
