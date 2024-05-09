package testonly

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type Controller interface {
	ResetState(ctx context.Context) error
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

func (c *ControllerImpl) ResetState(ctx context.Context) error {
	err := c.testOnlyService.TruncateTables(ctx)
	if err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	return nil
}
