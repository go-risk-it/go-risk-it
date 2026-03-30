package consumers

import (
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/lobby/api/messaging"
	"github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	"github.com/go-risk-it/go-risk-it/internal/lobby/logic/state"
)

// StateController translates between lobby state service results and messaging
// DTOs. It lives in consumers because its only caller is the broadcaster in this
// package.
type StateController struct {
	stateService state.Service
}

// NewStateController creates a StateController backed by the state service.
func NewStateController(
	stateService state.Service,
) *StateController {
	return &StateController{
		stateService: stateService,
	}
}

// GetLobbyState fetches the current lobby state and converts it to a messaging DTO.
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
