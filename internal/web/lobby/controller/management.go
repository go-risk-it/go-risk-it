package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/lobby/messaging"
	"github.com/go-risk-it/go-risk-it/internal/api/lobby/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/management"
)

type ManagementController interface {
	JoinLobby(ctx ctx.LobbyContext, request request.JoinLobby) error
	GetAvailableLobbies(ctx ctx.TraceContext) (messaging.Lobbies, error)
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

func (c *ManagementControllerImpl) GetAvailableLobbies(
	ctx ctx.TraceContext,
) (messaging.Lobbies, error) {
	lobbies, err := c.managementService.GetAvailableLobbies(ctx)
	if err != nil {
		return messaging.Lobbies{}, fmt.Errorf("failed to get available lobbies: %w", err)
	}

	return messaging.Lobbies{
		Lobbies: convertLobbies(lobbies),
	}, nil
}

func convertLobbies(cards []sqlc.GetAvailableLobbiesRow) []messaging.Lobby {
	result := make([]messaging.Lobby, len(cards))
	for i, c := range cards {
		result[i] = convertLobby(c)
	}

	return result
}

func convertLobby(lobby sqlc.GetAvailableLobbiesRow) messaging.Lobby {
	return messaging.Lobby{
		ID:                   lobby.ID,
		NumberOfParticipants: lobby.ParticipantCount,
	}
}
