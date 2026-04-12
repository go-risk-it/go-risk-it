package orchestration

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
)

func (s *orchestrator[T, R]) recordGameFinished(
	ctx gamectx.GameContext,
) {
	s.gameMetrics.ActiveGames.Add(ctx, -1)

	if elapsed, ok := s.gameTiming.ElapsedAndClear(ctx.GameID()); ok {
		s.gameMetrics.GameDuration.Record(ctx, elapsed.Seconds())
	}
}

// checkMission checks whether the current player's mission is accomplished
// using pure cached state. No DB queries — all data comes from ECST.
func (s *orchestrator[T, R]) checkMission(
	ctx gamectx.GameContext,
	prevState *snapshot.CachedGameState,
	effect moveservice.MoveEffect,
) (bool, error) {
	// Compute post-move regions — same pattern as IsDomination.
	postMoveRegions := ApplyRegionUpdates(
		prevState.PublicSnapshot.Regions,
		effect.RegionUpdates,
	)

	continents, err := s.boardService.GetContinents(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get continents for mission check: %w", err)
	}

	isMissionAccomplished, err := s.missionService.IsMissionAccomplished(
		ctx,
		postMoveRegions,
		prevState.PrivateSnapshots,
		continents,
	)
	if err != nil {
		return false, fmt.Errorf(
			"unable to check if mission is accomplished: %w", err,
		)
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
