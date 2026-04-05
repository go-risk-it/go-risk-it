package routes

import (
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/lobby/api/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/lobby/api/rest/response"
	"github.com/go-risk-it/go-risk-it/internal/lobby/internal/logic/creation"
)

type CreationController struct {
	creationService creation.Service
}

func NewCreationController(
	creationService creation.Service,
) *CreationController {
	return &CreationController{
		creationService: creationService,
	}
}

func (c *CreationController) CreateLobby(
	ctx ctx.UserContext,
	request request.CreateLobby,
) (response.CreateLobby, error) {
	lobbyID, err := c.creationService.CreateLobby(ctx, request.OwnerName)
	if err != nil {
		return response.CreateLobby{}, err
	}

	return response.CreateLobby{LobbyID: lobbyID}, nil
}
