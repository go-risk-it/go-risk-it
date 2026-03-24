package creation

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	dbutil "github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/db"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/metrics"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateLobby(ctx ctx.UserContext, ownerName string) (int64, error)
}

type ServiceImpl struct {
	querier db.Querier
	metrics *metrics.Metrics
}

var _ Service = (*ServiceImpl)(nil)

func NewService(querier db.Querier, m *metrics.Metrics) *ServiceImpl {
	return &ServiceImpl{
		querier: querier,
		metrics: m,
	}
}

func (s *ServiceImpl) CreateLobby(ctx ctx.UserContext, ownerName string) (int64, error) {
	lobbyID, err := dbutil.InTransaction(
		s.querier,
		ctx,
		s.metrics,
		func(qtx db.Querier) (int64, error) {
			return s.CreateLobbyQ(ctx, qtx, ownerName)
		},
	)
	if err != nil {
		return -1, fmt.Errorf("failed to create lobby: %w", err)
	}

	return lobbyID, nil
}

func (s *ServiceImpl) CreateLobbyQ(
	ctx ctx.UserContext,
	querier db.Querier,
	ownerName string,
) (int64, error) {
	slog.InfoContext(ctx, "creating lobby")

	lobbyID, err := querier.CreateLobby(ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to create lobby: %w", err)
	}

	slog.InfoContext(ctx, "lobby created", "lobbyID", lobbyID)

	participantID, err := querier.InsertParticipant(ctx, sqlc.InsertParticipantParams{
		LobbyID: lobbyID,
		UserID:  ctx.UserID(),
		Name:    ownerName,
	})
	if err != nil {
		return -1, fmt.Errorf("failed to insert participant: %w", err)
	}

	slog.InfoContext(ctx, "participant inserted", "participantID", participantID)

	if err := querier.UpdateLobbyOwner(ctx, sqlc.UpdateLobbyOwnerParams{
		OwnerID: pgtype.Int8{
			Int64: participantID,
			Valid: true,
		},
		ID: lobbyID,
	}); err != nil {
		return -1, fmt.Errorf("failed to update lobby owner: %w", err)
	}

	slog.InfoContext(ctx, "lobby owner updated")

	return lobbyID, nil
}
