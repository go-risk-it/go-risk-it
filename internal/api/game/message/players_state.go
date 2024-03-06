package message

type Player struct {
	ID             string `json:"id"`
	Index          int64  `json:"index"`
	TroopsToDeploy int64  `json:"troopsToDeploy"`
}

type PlayersState struct {
	Players []Player `json:"players"`
}
