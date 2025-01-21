package ctx

type UserContext interface {
	TraceContext
	UserID() string
}

type userContext struct {
	TraceContext
	userID string
}

var _ UserContext = (*userContext)(nil)

func (c *userContext) UserID() string {
	return c.userID
}

func WithUserID(ctx TraceContext, userID string) UserContext {
	ctx.SetLog(ctx.Log().With("userID", userID))

	return &userContext{
		TraceContext: ctx,
		userID:       userID,
	}
}
