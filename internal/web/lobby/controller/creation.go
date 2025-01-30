package controller

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/creation"
)

type CreationController interface {
	CreateLobby(ctx ctx.UserContext) (int64, error)
}

type CreationControllerImpl struct {
	creationService creation.Service
}

var _ CreationController = (*CreationControllerImpl)(nil)

func NewCreationController(
	creationService creation.Service,
) *CreationControllerImpl {
	return &CreationControllerImpl{
		creationService: creationService,
	}
}

func (c *CreationControllerImpl) CreateLobby(
	ctx ctx.UserContext,
) (int64, error) {
	return c.creationService.CreateLobby(ctx)
}
