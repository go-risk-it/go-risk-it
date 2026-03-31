package orchestration

import (
	"fmt"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
)

func (s *orchestrator[T, R]) recordGameFinished(
	ctx gamectx.GameContext,
) {
	s.gameMetrics.ActiveGames.Add(ctx, -1)

	if elapsed, ok := s.gameTiming.ElapsedAndClear(ctx.GameID()); ok {
		s.gameMetrics.GameDuration.Record(ctx, elapsed.Seconds())
	}
}

func (s *orchestrator[T, R]) checkMission(
	ctx gamectx.GameContext,
	querier db.Querier,
) (bool, error) {
	isMissionAccomplished, err := s.missionService.IsMissionAccomplished(
		ctx, querier,
	)
	if err != nil {
		return false, fmt.Errorf(
			"unable to check if mission is accomplished: %w", err,
		)
	}

	if isMissionAccomplished {
		observe.SpanEvent(ctx, "game_is_over")
		s.recordGameFinished(ctx)
	}

	return isMissionAccomplished, nil
}
