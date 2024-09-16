package ctx

type GameContext interface {
	UserContext
	GameID() int64
}

type gameContext struct {
	UserContext
	gameID int64
}

var _ GameContext = (*gameContext)(nil)

func (c *gameContext) GameID() int64 {
	return c.gameID
}

func WithGameID(ctx UserContext, gameID int64) GameContext {
	ctx.SetLog(ctx.Log().With("gameID", gameID))

	return &gameContext{
		UserContext: ctx,
		gameID:      gameID,
	}
}
