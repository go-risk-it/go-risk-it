package message

type Region struct {
	RegionID int64 `json:"regionId"`
	OwnerID  int64 `json:"ownerId"`
	Troops   int64 `json:"troops"`
}

type BoardState struct {
	Regions []Region `json:"regions"`
}
