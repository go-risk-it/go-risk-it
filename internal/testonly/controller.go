package testonly

import (
	"context"
	"fmt"
)

type Controller interface {
	ResetState(ctx context.Context) error
	SetupNearWin(ctx context.Context, gameID int64) error
}

type controller struct {
	testOnlyService Service
}

var _ Controller = (*controller)(nil)

func NewController(testOnlyService Service) Controller {
	return &controller{
		testOnlyService: testOnlyService,
	}
}

func (c *controller) ResetState(ctx context.Context) error {
	err := c.testOnlyService.TruncateTables(ctx)
	if err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	return nil
}

func (c *controller) SetupNearWin(ctx context.Context, gameID int64) error {
	err := c.testOnlyService.SetupNearWin(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to setup near win: %w", err)
	}

	return nil
}
