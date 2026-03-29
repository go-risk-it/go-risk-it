package routes

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/lobby/api/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/lobby/api/rest/response"
	"github.com/go-risk-it/go-risk-it/internal/lobby/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/lobby/logic/management"
)

type ManagementController struct {
	managementService management.Service
}

func NewManagementController(
	managementService management.Service,
) *ManagementController {
	return &ManagementController{
		managementService: managementService,
	}
}

func (c *ManagementController) JoinLobby(
	ctx ctx.LobbyContext,
	request request.JoinLobby,
) error {
	return c.managementService.JoinLobby(ctx, request.ParticipantName)
}

func (c *ManagementController) GetUserLobbies(ctx ctx.UserContext) (response.Lobbies, error) {
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
