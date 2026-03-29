package state

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/tracing"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
)

type Game struct {
	ID           int64
	Turn         int64
	Phase        sqlc.GamePhaseType
	WinnerUserID string
}

type Service interface {
	GetGameState(ctx ctx.GameContext) (*Game, error)
	GetGameStateWithQuerier(ctx ctx.GameContext, querier db.Querier) (*Game, error)
	GetUserGames(ctx kernelctx.UserContext) ([]int64, error)
}

type service struct {
	querier db.Querier
}

var _ Service = (*service)(nil)

func NewService(
	querier db.Querier,
) Service {
	return &service{
		querier: querier,
	}
}

func (s *service) GetGameState(ctx ctx.GameContext) (*Game, error) {
	return s.GetGameStateWithQuerier(ctx, s.querier)
}

func (s *service) GetGameStateWithQuerier(ctx ctx.GameContext, querier db.Querier) (*Game, error) {
	ctx, span := tracing.StartGameSpan(ctx, "game.advance.get_state")
	defer span.End()

	game, err := querier.GetGame(ctx, ctx.GameID())
	if err != nil {
		slog.WarnContext(ctx, "failed to get game", "error", err)

		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	winnerUserID := ""
	if game.WinnerUserID.Valid {
		winnerUserID = game.WinnerUserID.String
	}

	return &Game{
		ID:           game.ID,
		Turn:         game.Turn,
		Phase:        game.CurrentPhase,
		WinnerUserID: winnerUserID,
	}, nil
}

func (s *service) GetUserGames(ctx kernelctx.UserContext) ([]int64, error) {
	slog.InfoContext(ctx, "getting user games")

	userGames, err := s.querier.GetUserGames(ctx, ctx.UserID())
	if err != nil {
		return nil, fmt.Errorf("failed to get joined games: %w", err)
	}

	slog.InfoContext(ctx, "got user games", "games", userGames)

	return userGames, nil
}
