package orchestration

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
)

// getOrWarmPrevState returns the cached state for the game, warming from the
// database on a cache miss. The warm path uses the transactional querier so
// that the read is consistent with any preceding writes in the same TX.
func (s *orchestrator[T, R]) getOrWarmPrevState(
	ctx gamectx.GameContext,
	querier db.Querier,
	turn int64,
) (*snapshot.CachedGameState, error) {
	if cached := s.stateStore.Get(ctx.GameID()); cached != nil {
		return cached, nil
	}

	// Cache miss — warm from DB using the transactional querier.
	reader := s.snapshotReaderFactory(querier)

	publicSnapshot, err := reader.GetPublicSnapshot(ctx)
	if err != nil {
		return nil, fmt.Errorf("warming public snapshot: %w", err)
	}

	privateSnapshots, err := reader.GetAllPrivateSnapshots(ctx)
	if err != nil {
		return nil, fmt.Errorf("warming private snapshots: %w", err)
	}

	conquered, err := querier.HasConqueredInTurn(ctx, sqlc.HasConqueredInTurnParams{
		ID:   ctx.GameID(),
		Turn: turn,
	})
	if err != nil {
		return nil, fmt.Errorf("checking conquered in turn: %w", err)
	}

	state := &snapshot.CachedGameState{
		Turn:             turn,
		ConqueredInTurn:  conquered,
		PublicSnapshot:   publicSnapshot,
		PrivateSnapshots: privateSnapshots,
	}

	s.stateStore.Store(ctx.GameID(), state)

	return state, nil
}

// buildWalkContext constructs a WalkContext from the previous cached state and
// the move effect, ready for the phase walker.
func buildWalkContext(
	prevState *snapshot.CachedGameState,
	effect moveservice.MoveEffect,
	voluntary bool,
	currentUserID string,
) moveservice.WalkContext {
	return moveservice.WalkContext{
		Voluntary:        voluntary,
		PrevSnapshot:     prevState.PublicSnapshot,
		PrivateSnapshots: prevState.PrivateSnapshots,
		Effect:           effect,
		CurrentUserID:    currentUserID,
	}
}

// buildAdvanceContext constructs an AdvanceContext from the previous cached
// state, the move effect, and the board's continent definitions.
func (s *orchestrator[T, R]) buildAdvanceContext(
	ctx gamectx.GameContext,
	prevState *snapshot.CachedGameState,
	effect moveservice.MoveEffect,
	currentUserID string,
) (moveservice.AdvanceContext, error) {
	continents, err := s.boardService.GetContinents(ctx)
	if err != nil {
		return moveservice.AdvanceContext{}, fmt.Errorf("getting continents: %w", err)
	}

	// Compute the updated regions by applying the move effect to the previous snapshot.
	updatedRegions := ApplyRegionUpdates(prevState.PublicSnapshot.Regions, effect.RegionUpdates)

	return moveservice.AdvanceContext{
		ConqueredInTurn: prevState.ConqueredInTurn,
		UpdatedRegions:  updatedRegions,
		CurrentUserID:   currentUserID,
		Continents:      continents,
	}, nil
}
