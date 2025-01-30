package ctx

type LobbyContext interface {
	UserContext
	LobbyID() int64
}

type lobbyContext struct {
	UserContext
	lobbyID int64
}

var _ LobbyContext = (*lobbyContext)(nil)

func (c *lobbyContext) LobbyID() int64 {
	return c.lobbyID
}

func WithLobbyID(ctx UserContext, lobbyID int64) LobbyContext {
	ctx.SetLog(ctx.Log().With("lobbyID", lobbyID))

	return &lobbyContext{
		UserContext: ctx,
		lobbyID:     lobbyID,
	}
}
