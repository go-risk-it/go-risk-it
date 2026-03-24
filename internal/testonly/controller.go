package testonly

import (
	"context"
	"fmt"
)

type Controller interface {
	ResetState(ctx context.Context) error
	SetupNearWin(ctx context.Context, gameID int64) error
}

type ControllerImpl struct {
	testOnlyService Service
}

var _ Controller = (*ControllerImpl)(nil)

func NewController(testOnlyService Service) *ControllerImpl {
	return &ControllerImpl{
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

func (c *ControllerImpl) SetupNearWin(ctx context.Context, gameID int64) error {
	err := c.testOnlyService.SetupNearWin(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to setup near win: %w", err)
	}

	return nil
}
