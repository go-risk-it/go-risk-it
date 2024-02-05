package response

type Region struct {
	RegionID int64  `json:"regionId"`
	Name     string `json:"name"`
	Owner    int64  `json:"owner"`
	Troops   int64  `json:"troops"`
}

type BoardStateResponse struct {
	Regions []Region `json:"regions"`
}
