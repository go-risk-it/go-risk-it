package player

// Player is the domain type for a player joining a game.
// API-layer types (e.g., request.Player) should be mapped to this at the controller boundary.
type Player struct {
	UserID string
	Name   string
}
