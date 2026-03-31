package creation

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	dbutil "github.com/go-risk-it/go-risk-it/internal/kernel/data"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/go-risk-it/go-risk-it/internal/lobby/data/db"
	"github.com/go-risk-it/go-risk-it/internal/lobby/data/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateLobby(ctx ctx.UserContext, ownerName string) (int64, error)
	CreateLobbyWithQuerier(
		ctx ctx.UserContext,
		querier db.Querier,
		ownerName string,
	) (int64, error)
}

type service struct {
	querier db.Querier
	metrics *metrics.StateMetrics
}

var _ Service = (*service)(nil)

func NewService(querier db.Querier, m *metrics.StateMetrics) Service {
	return &service{
		querier: querier,
		metrics: m,
	}
}

func (s *service) CreateLobby(ctx ctx.UserContext, ownerName string) (int64, error) {
	lobbyID, err := dbutil.InTransaction(
		s.querier,
		ctx,
		s.metrics,
		func(qtx db.Querier) (int64, error) {
			return s.CreateLobbyWithQuerier(ctx, qtx, ownerName)
		},
	)
	if err != nil {
		return -1, fmt.Errorf("failed to create lobby: %w", err)
	}

	return lobbyID, nil
}

func (s *service) CreateLobbyWithQuerier(
	ctx ctx.UserContext,
	querier db.Querier,
	ownerName string,
) (int64, error) {
	lobbyID, err := querier.CreateLobby(ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to create lobby: %w", err)
	}

	participantID, err := querier.InsertParticipant(ctx, sqlc.InsertParticipantParams{
		LobbyID: lobbyID,
		UserID:  ctx.UserID(),
		Name:    ownerName,
	})
	if err != nil {
		return -1, fmt.Errorf("failed to insert participant: %w", err)
	}

	if err := querier.UpdateLobbyOwner(ctx, sqlc.UpdateLobbyOwnerParams{
		OwnerID: pgtype.Int8{
			Int64: participantID,
			Valid: true,
		},
		ID: lobbyID,
	}); err != nil {
		return -1, fmt.Errorf("failed to update lobby owner: %w", err)
	}

	return lobbyID, nil
}
