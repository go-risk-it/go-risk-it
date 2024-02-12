package response

type Region struct {
	RegionID int64 `json:"regionId"`
	OwnerID  int64 `json:"ownerId"`
	Troops   int64 `json:"troops"`
}

type BoardStateResponse struct {
	Regions []Region `json:"regions"`
}
