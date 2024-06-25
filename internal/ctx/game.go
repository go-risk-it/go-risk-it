package ctx

type GameContext interface {
	LogContext
	GameID() int64
}

type gameContext struct {
	LogContext
	gameID int64
}

var _ GameContext = (*gameContext)(nil)

func (c *gameContext) GameID() int64 {
	return c.gameID
}

func WithGameID(ctx LogContext, gameID int64) GameContext {
	return &gameContext{
		LogContext: ctx,
		gameID:     gameID,
	}
}
