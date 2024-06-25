package ctx

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

type MoveContext interface {
	UserContext
	GameContext
}

type moveContext struct {
	UserContext
	GameContext
}

var _ MoveContext = (*moveContext)(nil)

func NewMoveContext(userCtx UserContext, gameCtx GameContext) MoveContext {
	return &moveContext{
		UserContext: userCtx,
		GameContext: gameCtx,
	}
}

func (c *moveContext) Log() *zap.SugaredLogger {
	return c.GameContext.Log()
}

func (c *moveContext) GameID() int64 {
	return c.GameContext.GameID()
}

func (c *moveContext) UserID() string {
	return c.UserContext.UserID()
}

func (c *moveContext) Deadline() (time.Time, bool) {
	return c.GameContext.Deadline()
}

func (c *moveContext) Done() <-chan struct{} {
	return c.GameContext.Done()
}

func (c *moveContext) Err() error {
	return fmt.Errorf("move context: %w", c.GameContext.Err())
}

func (c *moveContext) Value(key any) any {
	return c.GameContext.Value(key)
}
