package request

type BoardStateRequest struct {
	GameID int64 `json:"gameId"`
	UserID int64 `json:"userId"`
}
