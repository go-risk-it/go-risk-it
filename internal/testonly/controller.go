package testonly

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"go.uber.org/zap"
)

type Controller interface {
	ResetState(ctx ctx.LogContext) error
}

type ControllerImpl struct {
	log             *zap.SugaredLogger
	testOnlyService Service
}

func NewController(
	log *zap.SugaredLogger,
	testOnlyService Service,
) *ControllerImpl {
	return &ControllerImpl{
		log:             log,
		testOnlyService: testOnlyService,
	}
}

func (c *ControllerImpl) ResetState(ctx ctx.LogContext) error {
	err := c.testOnlyService.TruncateTables(ctx)
	if err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	return nil
}
