package riskcontext

import "context"

type UserContext interface {
	context.Context
	UserID() string
}

type GameContext interface {
	context.Context
	GameID() int64
}

type MoveContext interface {
	UserContext
	GameContext
}

type userContext struct {
	context.Context
	userID string
}

func (c *userContext) UserID() string {
	return c.userID
}

func WithUserID(ctx context.Context, userID string) UserContext {
	return &userContext{
		Context: ctx,
		userID:  userID,
	}
}

type moveContext struct {
	UserContext
	gameID int64
}

func (c *moveContext) GameID() int64 {
	return c.gameID
}

func WithGameID(ctx UserContext, gameID int64) MoveContext {
	return &moveContext{
		UserContext: ctx,
		gameID:      gameID,
	}
}
