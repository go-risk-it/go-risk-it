package message

type FullState struct {
	GameState  GameState  `json:"gameState"`
	BoardState BoardState `json:"boardState"`
}
