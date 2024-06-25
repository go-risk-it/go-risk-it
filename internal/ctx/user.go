package ctx

type UserContext interface {
	LogContext
	UserID() string
}

type userContext struct {
	LogContext
	userID string
}

var _ UserContext = (*userContext)(nil)

func (c *userContext) UserID() string {
	return c.userID
}

func WithUserID(ctx LogContext, userID string) UserContext {
	return &userContext{
		LogContext: ctx,
		userID:     userID,
	}
}
