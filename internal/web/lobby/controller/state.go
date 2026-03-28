package controller

import (
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/api/lobby/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/state"
)

type StateController struct {
	stateService state.Service
}

func NewStateController(
	stateService state.Service,
) *StateController {
	return &StateController{
		stateService: stateService,
	}
}

func (s *StateController) GetLobbyState(ctx ctx.LobbyContext) (messaging.LobbyState, error) {
	lobby, err := s.stateService.GetLobbyState(ctx)
	if err != nil {
		slog.WarnContext(ctx, "failed to get lobby state", "error", err)

		return messaging.LobbyState{}, err
	}

	return messaging.LobbyState{
		ID:           lobby.ID,
		Participants: convertParticipants(lobby.Participants),
	}, nil
}

func convertParticipants(participants []state.Participant) []messaging.Participant {
	result := make([]messaging.Participant, 0, len(participants))
	for _, participant := range participants {
		result = append(result, messaging.Participant{
			UserID: participant.UserID,
		})
	}

	return result
}
