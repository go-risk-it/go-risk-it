package testonly

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
)

type Controller interface {
	ResetState(ctx ctx.LogContext) error
	SetupNearWin(ctx ctx.LogContext, gameID int64) error
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

func (c *ControllerImpl) ResetState(ctx ctx.LogContext) error {
	err := c.testOnlyService.TruncateTables(ctx)
	if err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	return nil
}

func (c *ControllerImpl) SetupNearWin(ctx ctx.LogContext, gameID int64) error {
	err := c.testOnlyService.SetupNearWin(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to setup near win: %w", err)
	}

	return nil
}
