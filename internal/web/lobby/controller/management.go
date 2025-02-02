package controller

import (
	"github.com/go-risk-it/go-risk-it/internal/api/lobby/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/management"
)

type ManagementController interface {
	JoinLobby(ctx ctx.LobbyContext, request request.JoinLobby) error
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

func (c *ManagementControllerImpl) JoinLobby(
	ctx ctx.LobbyContext,
	request request.JoinLobby,
) error {
	return c.managementService.JoinLobby(ctx, request.ParticipantName)
}
