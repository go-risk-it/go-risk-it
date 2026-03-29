package controller

import (
	"github.com/go-risk-it/go-risk-it/internal/api/lobby/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/creation"
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
) (int64, error) {
	return c.creationService.CreateLobby(ctx, request.OwnerName)
}
