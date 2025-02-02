package controller

import (
	"github.com/go-risk-it/go-risk-it/internal/api/lobby/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/state"
)

type StateController interface {
	GetLobbyState(ctx ctx.LobbyContext) (messaging.LobbyState, error)
}

type StateControllerImpl struct {
	stateService state.Service
}

var _ StateController = (*StateControllerImpl)(nil)

func NewStateController(
	stateService state.Service,
) *StateControllerImpl {
	return &StateControllerImpl{
		stateService: stateService,
	}
}

func (s *StateControllerImpl) GetLobbyState(ctx ctx.LobbyContext) (messaging.LobbyState, error) {
	lobby, err := s.stateService.GetLobbyState(ctx)
	if err != nil {
		ctx.Log().Warnw("failed to get lobby state", "err", err)
	}

	return messaging.LobbyState{
		ID:           lobby.ID,
		Participants: convertParticipants(lobby.Participants),
	}, nil
}

func convertParticipants(participants []state.Participant) []messaging.Participant {
	result := make([]messaging.Participant, 0)
	for _, participant := range participants {
		result = append(result, messaging.Participant{
			UserID: participant.UserID,
		})
	}

	return result
}
