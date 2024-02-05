package request

type GameStateRequest struct {
	UserID int64 `json:"userId"`
	GameID int64 `json:"gameId"`
}
