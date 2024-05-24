package request

type Player struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
}

type CreateGame struct {
	Players []Player `json:"players"`
}
