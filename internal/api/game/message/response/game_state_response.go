package response

type User struct {
	UserID    int64 `json:"userId"`
	TurnIndex int64 `json:"turnIndex"`
}
type GameStateResponse struct {
	GameID      int64  `json:"gameId"`
	Users       []User `json:"users"`
	CurrentTurn int64  `json:"currentTurn"`
}
