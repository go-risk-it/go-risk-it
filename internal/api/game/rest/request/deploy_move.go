package request

type DeployMove struct {
	PlayerID string `json:"playerId"`
	RegionID string `json:"regionId"`
	Troops   int    `json:"troops"`
}
