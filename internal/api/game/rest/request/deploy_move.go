package request

type DeployMove struct {
	UserID   string `json:"userId"`
	RegionID string `json:"regionId"`
	Troops   int    `json:"troops"`
}
