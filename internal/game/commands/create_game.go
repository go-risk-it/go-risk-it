package commands

import "context"

// Player identifies a participant in a game creation command.
type Player struct {
	UserID string
	Name   string
}

// CreateGame is the cross-module command for creating a new game from a lobby.
type CreateGame struct {
	Players []Player
}

// CreateGameResult is the response from a successful game creation.
type CreateGameResult struct {
	GameID int64
}

// Handler processes game commands dispatched by the kernel router.
type Handler interface {
	HandleCreateGame(ctx context.Context, cmd CreateGame) (CreateGameResult, error)
}
