package ctx

import (
	"context"

	"go.uber.org/zap"
)

type LogContext interface {
	context.Context
	Log() *zap.SugaredLogger
}

type logContext struct {
	context.Context
	log *zap.SugaredLogger
}

var _ LogContext = (*logContext)(nil)

func (c *logContext) Log() *zap.SugaredLogger {
	return c.log
}

func WithLog(ctx context.Context, log *zap.SugaredLogger) LogContext {
	return &logContext{
		Context: ctx,
		log:     log,
	}
}
