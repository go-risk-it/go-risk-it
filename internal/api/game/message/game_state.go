package message

type GameState struct {
	GameID           int64  `json:"gameId"`
	CurrentTurn      int64  `json:"currentTurn"`
	CurrentPhase     string `json:"currentPhase"`
	DeployableTroops int64  `json:"deployableTroops"`
}
