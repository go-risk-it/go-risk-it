package state

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
)

type Game struct {
	ID           int64
	Turn         int64
	Phase        sqlc.PhaseType
	WinnerUserID string
}

type Service interface {
	GetGameState(ctx ctx.GameContext) (*Game, error)
	GetGameStateQ(ctx ctx.GameContext, querier db.Querier) (*Game, error)
	GetUserGames(ctx ctx.UserContext) ([]int64, error)
}

type ServiceImpl struct {
	querier db.Querier
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	querier db.Querier,
) *ServiceImpl {
	return &ServiceImpl{
		querier: querier,
	}
}

func (s *ServiceImpl) GetGameState(ctx ctx.GameContext) (*Game, error) {
	return s.GetGameStateQ(ctx, s.querier)
}

func (s *ServiceImpl) GetGameStateQ(ctx ctx.GameContext, querier db.Querier) (*Game, error) {
	game, err := querier.GetGame(ctx, ctx.GameID())
	if err != nil {
		ctx.Log().Warnw("failed to get game", "err", err)

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

func (s *ServiceImpl) GetUserGames(ctx ctx.UserContext) ([]int64, error) {
	ctx.Log().Infow("getting user games")

	userGames, err := s.querier.GetUserGames(ctx, ctx.UserID())
	if err != nil {
		return nil, fmt.Errorf("failed to get joined games: %w", err)
	}

	ctx.Log().Infow("got user games", "games", userGames)

	return userGames, nil
}
