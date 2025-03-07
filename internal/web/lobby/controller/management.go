package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/lobby/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/api/lobby/rest/response"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/management"
)

type ManagementController interface {
	JoinLobby(ctx ctx.LobbyContext, request request.JoinLobby) error
	GetUserLobbies(ctx ctx.UserContext) (response.Lobbies, error)
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

func (c *ManagementControllerImpl) GetUserLobbies(ctx ctx.UserContext) (response.Lobbies, error) {
	userLobbies, err := c.managementService.GetUserLobbies(ctx)
	if err != nil {
		return response.Lobbies{}, fmt.Errorf("failed to get available lobbies: %w", err)
	}

	return response.Lobbies{
		Owned:    convertToLobbies(userLobbies.Owned),
		Joined:   convertToLobbies(userLobbies.Joined),
		Joinable: convertToLobbies(userLobbies.Joinable),
	}, nil
}

func convertToLobbies[T any](rows []T) []response.Lobby {
	res := make([]response.Lobby, len(rows))

	for idx, row := range rows {
		r := any(row)
		switch lobby := r.(type) {
		case sqlc.GetOwnedLobbiesRow:
			res[idx] = response.Lobby{ID: lobby.ID, NumberOfParticipants: lobby.ParticipantCount}
		case sqlc.GetJoinedLobbiesRow:
			res[idx] = response.Lobby{ID: lobby.ID, NumberOfParticipants: lobby.ParticipantCount}
		case sqlc.GetJoinableLobbiesRow:
			res[idx] = response.Lobby{ID: lobby.ID, NumberOfParticipants: lobby.ParticipantCount}
		}
	}

	return res
}
