package message

type Player struct {
	ID    string `json:"id"`
	Index int64  `json:"index"`
}

type PlayersState struct {
	Players []Player `json:"players"`
}
