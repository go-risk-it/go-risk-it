package response

// StartGame is the response body for a successful lobby start.
type StartGame struct {
	GameID int64 `json:"gameId"`
}
