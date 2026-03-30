package orchestration

import (
	"fmt"
	"log/slog"
	"time"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func (s *orchestrator[T, R]) recordMetrics(
	ctx gamectx.GameContext,
	start time.Time,
) {
	phase := string(s.service.PhaseType())
	phaseAttr := metric.WithAttributes(
		attribute.String("phase", phase),
	)

	s.gameMetrics.MovesTotal.Add(ctx, 1, phaseAttr)
	s.gameMetrics.PhaseDuration.Record(
		ctx, time.Since(start).Seconds(), phaseAttr,
	)
}

func (s *orchestrator[T, R]) recordGameFinished(
	ctx gamectx.GameContext,
) {
	s.gameMetrics.GamesFinished.Add(ctx, 1)
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
		slog.InfoContext(ctx, "game is over")
		s.recordGameFinished(ctx)
	}

	return isMissionAccomplished, nil
}
