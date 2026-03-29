package creation

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/data/lobby/db"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	dbutil "github.com/go-risk-it/go-risk-it/internal/kernel/data"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
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
	metrics *metrics.InfraMetrics
}

var _ Service = (*service)(nil)

func NewService(querier db.Querier, m *metrics.InfraMetrics) Service {
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
