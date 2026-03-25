package start

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/db"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CanStartLobby(ctx ctx.LobbyContext) (bool, error)
	GetLobbyPlayers(ctx ctx.LobbyContext) ([]sqlc.GetLobbyPlayersRow, error)
	MarkLobbyAsStarted(ctx ctx.LobbyContext, gameID int64) error
}

type service struct {
	querier db.Querier
}

var _ Service = (*service)(nil)

func NewService(querier db.Querier) Service {
	return &service{
		querier: querier,
	}
}

func (s *service) CanStartLobby(ctx ctx.LobbyContext) (bool, error) {
	slog.InfoContext(ctx, "checking if lobby can be started")

	canStartLobby, err := s.querier.CanLobbyBeStarted(ctx, sqlc.CanLobbyBeStartedParams{
		LobbyID:             ctx.LobbyID(),
		UserID:              ctx.UserID(),
		MinimumParticipants: 3,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check if lobby can be started: %w", err)
	}

	return canStartLobby, nil
}

func (s *service) GetLobbyPlayers(ctx ctx.LobbyContext) ([]sqlc.GetLobbyPlayersRow, error) {
	slog.DebugContext(ctx, "getting lobby players")

	lobbyPlayers, err := s.querier.GetLobbyPlayers(ctx, ctx.LobbyID())
	if err != nil {
		return nil, fmt.Errorf("failed to get lobby players: %w", err)
	}

	return lobbyPlayers, nil
}

func (s *service) MarkLobbyAsStarted(ctx ctx.LobbyContext, gameID int64) error {
	slog.DebugContext(ctx, "marking lobby as started")

	if err := s.querier.MarkLobbyAsStarted(ctx, sqlc.MarkLobbyAsStartedParams{
		LobbyID: ctx.LobbyID(),
		GameID: pgtype.Int8{
			Int64: gameID,
			Valid: true,
		},
	}); err != nil {
		return fmt.Errorf("failed to mark lobby as started: %w", err)
	}

	return nil
}
