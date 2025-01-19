package state

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
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
