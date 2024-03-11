package request

type DeployMove struct {
	GameID   int64
	PlayerID string
	RegionID string
	Troops   int
}
