package request

type FullStateRequest struct {
	UserID int64 `json:"userId"`
	GameID int64 `json:"gameId"`
}
