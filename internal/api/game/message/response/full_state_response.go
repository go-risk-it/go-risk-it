package response

type FullStateResponse struct {
	GameState  GameStateResponse  `json:"gameState"`
	BoardState BoardStateResponse `json:"boardState"`
}
