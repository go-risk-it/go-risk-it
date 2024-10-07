package messaging

type Player struct {
	UserID    string `json:"userId"`
	Name      string `json:"name"`
	Index     int64  `json:"index"`
	CardCount int64  `json:"cardCount"`
}

type PlayersState struct {
	Players []Player `json:"players"`
}
