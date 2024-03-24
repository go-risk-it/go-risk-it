package request

type DeployMove struct {
	GameID   int64  `json:"gameId"`
	PlayerID string `json:"playerId"`
	RegionID string `json:"regionId"`
	Troops   int    `json:"troops"`
}
