package messaging

type Region struct {
	ID      string `json:"id"`
	OwnerID string `json:"ownerId"`
	Troops  int64  `json:"troops"`
}

type BoardState struct {
	Regions []Region `json:"regions"`
}
