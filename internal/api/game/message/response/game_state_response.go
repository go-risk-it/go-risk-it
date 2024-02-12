package response

type Player struct {
	PlayerID  string `json:"playerId"`
	TurnIndex int64  `json:"turnIndex"`
}
type GameStateResponse struct {
	GameID       int64    `json:"gameId"`
	Players      []Player `json:"players"`
	CurrentTurn  int64    `json:"currentTurn"`
	CurrentPhase string   `json:"currentPhase"`
}
