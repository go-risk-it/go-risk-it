package controller

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/management"
)

type ManagementController interface {
	JoinLobby(ctx ctx.LobbyContext) error
}

type ManagementControllerImpl struct {
	managementService management.Service
}

var _ ManagementController = (*ManagementControllerImpl)(nil)

func NewManagementController(
	managementService management.Service,
) *ManagementControllerImpl {
	return &ManagementControllerImpl{
		managementService: managementService,
	}
}

func (c *ManagementControllerImpl) JoinLobby(ctx ctx.LobbyContext) error {
	return c.managementService.JoinLobby(ctx)
}
