package management

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/db"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/signals"
)

type Service interface {
	JoinLobby(ctx ctx.LobbyContext) error
}

type ServiceImpl struct {
	querier                 db.Querier
	lobbyStateChangedSignal signals.LobbyStateChangedSignal
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	querier db.Querier,
	lobbyStateChangedSignal signals.LobbyStateChangedSignal,
) *ServiceImpl {
	return &ServiceImpl{
		querier:                 querier,
		lobbyStateChangedSignal: lobbyStateChangedSignal,
	}
}

func (s *ServiceImpl) JoinLobby(ctx ctx.LobbyContext) error {
	if _, err := s.querier.ExecuteInTransaction(ctx, func(qtx db.Querier) (interface{}, error) {
		return nil, s.JoinLobbyQ(ctx, qtx)
	}); err != nil {
		return fmt.Errorf("failed to join lobby: %w", err)
	}

	go s.lobbyStateChangedSignal.Emit(ctx, signals.LobbyStateChangedData{})

	return nil
}

func (s *ServiceImpl) JoinLobbyQ(
	ctx ctx.LobbyContext,
	querier db.Querier,
) error {
	ctx.Log().Infow("joining lobby")

	participantID, err := querier.InsertParticipant(ctx, sqlc.InsertParticipantParams{
		LobbyID: ctx.LobbyID(),
		UserID:  ctx.UserID(),
	})
	if err != nil {
		return fmt.Errorf("failed to insert participant: %w", err)
	}

	ctx.Log().Infow("participant joined", "participant_id", participantID)

	return nil
}
