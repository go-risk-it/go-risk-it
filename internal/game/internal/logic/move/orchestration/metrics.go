package orchestration

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/jackc/pgx/v5/pgtype"
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

// IsDomination checks whether the current player owns all regions after the
// move effect is applied. This uses ECST data (prevState + effect) — no DB
// query needed. This is the universal win condition in Risk: controlling all
// territories wins regardless of the player's specific mission.
func IsDomination(
	prevState *snapshot.CachedGameState,
	effect moveservice.MoveEffect,
	userID string,
) bool {
	// Build ownership map from cached state.
	owners := make(map[string]string, len(prevState.PublicSnapshot.Regions))
	for _, r := range prevState.PublicSnapshot.Regions {
		owners[r.ID] = r.OwnerID
	}

	// Apply effect updates (region conquests change ownership).
	for _, u := range effect.RegionUpdates {
		owners[u.RegionID] = u.NewOwner
	}

	// Check if all regions belong to the current player.
	for _, owner := range owners {
		if owner != userID {
			return false
		}
	}

	return true
}

// assignWinner records the domination victory in the DB and emits metrics.
func (s *orchestrator[T, R]) assignWinner(
	ctx gamectx.GameContext,
	querier db.Querier,
) error {
	players, err := querier.GetPlayersByGame(ctx, ctx.GameID())
	if err != nil {
		return fmt.Errorf("failed to get players: %w", err)
	}

	for _, p := range players {
		if p.UserID == ctx.UserID() {
			if err := querier.AssignGameWinner(ctx, sqlc.AssignGameWinnerParams{
				WinnerPlayerID: pgtype.Int8{Int64: p.ID, Valid: true},
				GameID:         ctx.GameID(),
			}); err != nil {
				return fmt.Errorf("failed to assign game winner: %w", err)
			}

			break
		}
	}

	observe.SpanEvent(ctx, "domination_victory")
	s.recordGameFinished(ctx)

	return nil
}
