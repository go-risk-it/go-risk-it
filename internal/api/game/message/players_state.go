package message

type Player struct {
	UserID         string `json:"userId"`
	Name           string `json:"name"`
	Index          int64  `json:"index"`
	TroopsToDeploy int64  `json:"troopsToDeploy"`
}

type PlayersState struct {
	Players []Player `json:"players"`
}
