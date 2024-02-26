package message

type Region struct {
	ID      string `json:"id"`
	OwnerID int64  `json:"ownerId"`
	Troops  int64  `json:"troops"`
}

type BoardState struct {
	Regions []Region `json:"regions"`
}
