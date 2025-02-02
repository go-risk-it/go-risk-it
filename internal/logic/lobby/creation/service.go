package creation

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/db"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateLobby(ctx ctx.UserContext, ownerName string) (int64, error)
}

type ServiceImpl struct {
	querier db.Querier
}

var _ Service = (*ServiceImpl)(nil)

func NewService(querier db.Querier) *ServiceImpl {
	return &ServiceImpl{
		querier: querier,
	}
}

func (s *ServiceImpl) CreateLobby(ctx ctx.UserContext, ownerName string) (int64, error) {
	lobbyID, err := s.querier.ExecuteInTransaction(ctx, func(qtx db.Querier) (interface{}, error) {
		return s.CreateLobbyQ(ctx, qtx, ownerName)
	})
	if err != nil {
		return -1, fmt.Errorf("failed to create lobby: %w", err)
	}

	lobbyIDInt, ok := lobbyID.(int64)
	if !ok {
		return -1, fmt.Errorf("failed to convert lobbyID to int64: %w", err)
	}

	return lobbyIDInt, nil
}

func (s *ServiceImpl) CreateLobbyQ(
	ctx ctx.UserContext,
	querier db.Querier,
	ownerName string,
) (int64, error) {
	ctx.Log().Infow("creating lobby")

	lobbyID, err := querier.CreateLobby(ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to create lobby: %w", err)
	}

	ctx.Log().Infow("lobby created", "lobbyID", lobbyID)

	participantID, err := querier.InsertParticipant(ctx, sqlc.InsertParticipantParams{
		LobbyID: lobbyID,
		UserID:  ctx.UserID(),
		Name:    ownerName,
	})
	if err != nil {
		return -1, fmt.Errorf("failed to insert participant: %w", err)
	}

	ctx.Log().Infow("participant inserted", "participantID", participantID)

	if err := querier.UpdateLobbyOwner(ctx, sqlc.UpdateLobbyOwnerParams{
		OwnerID: pgtype.Int8{
			Int64: participantID,
			Valid: true,
		},
		ID: lobbyID,
	}); err != nil {
		return -1, fmt.Errorf("failed to update lobby owner: %w", err)
	}

	ctx.Log().Infow("lobby owner updated")

	return lobbyID, nil
}
