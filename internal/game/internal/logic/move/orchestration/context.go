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
// database on a cache miss. The warm path uses the provided querier directly
// (no transaction needed — per-game mutex provides serialization).
func (s *orchestrator[T, R]) getOrWarmPrevState(
	ctx gamectx.GameContext,
	querier db.Querier,
	turn int64,
) (*snapshot.CachedGameState, error) {
	if cached := s.stateStore.Get(ctx.GameID()); cached != nil {
		return s.ensureDeckWarmed(ctx, querier, cached)
	}

	return s.warmFullState(ctx, querier, turn)
}

// ensureDeckWarmed returns the cached state, warming just the deck from the DB
// if it was populated without one (e.g., older creation code).
func (s *orchestrator[T, R]) ensureDeckWarmed(
	ctx gamectx.GameContext,
	querier db.Querier,
	cached *snapshot.CachedGameState,
) (*snapshot.CachedGameState, error) {
	if cached.AvailableDeck != nil {
		return cached, nil
	}

	deckRows, err := querier.GetAvailableDeck(ctx, ctx.GameID())
	if err != nil {
		return nil, fmt.Errorf("warming available deck for cached state: %w", err)
	}

	deck, err := mapDeckRows(deckRows)
	if err != nil {
		return nil, fmt.Errorf("mapping available deck for cached state: %w", err)
	}

	cached.AvailableDeck = deck

	return cached, nil
}

// warmFullState reads the full game state from the database on a cache miss.
func (s *orchestrator[T, R]) warmFullState(
	ctx gamectx.GameContext,
	querier db.Querier,
	turn int64,
) (*snapshot.CachedGameState, error) {
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

	deckRows, err := querier.GetAvailableDeck(ctx, ctx.GameID())
	if err != nil {
		return nil, fmt.Errorf("warming available deck: %w", err)
	}

	deck, err := mapDeckRows(deckRows)
	if err != nil {
		return nil, fmt.Errorf("mapping available deck: %w", err)
	}

	state := &snapshot.CachedGameState{
		Turn:             turn,
		ConqueredInTurn:  conquered,
		AvailableDeck:    deck,
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
		Turn:            prevState.Turn,
		CurrentPhase:    string(prevState.PublicSnapshot.Phase.Type),
		Players:         prevState.PublicSnapshot.Players,
		AvailableDeck:   prevState.AvailableDeck,
	}, nil
}

// mapDeckRows converts sqlc query rows to snapshot CardState values.
func mapDeckRows(rows []sqlc.GetAvailableDeckRow) ([]snapshot.CardState, error) {
	deck := make([]snapshot.CardState, 0, len(rows))

	for _, row := range rows {
		cardType, err := mapSqlcCardType(row.CardType)
		if err != nil {
			return nil, err
		}

		region := ""
		if row.Region.Valid {
			region = row.Region.String
		}

		deck = append(deck, snapshot.CardState{
			ID:     row.ID,
			Type:   cardType,
			Region: region,
		})
	}

	return deck, nil
}

// mapSqlcCardType converts a sqlc GameCardType to a snapshot CardType.
func mapSqlcCardType(cardType sqlc.GameCardType) (snapshot.CardType, error) {
	switch cardType {
	case sqlc.GameCardTypeINFANTRY:
		return snapshot.CardInfantry, nil
	case sqlc.GameCardTypeCAVALRY:
		return snapshot.CardCavalry, nil
	case sqlc.GameCardTypeARTILLERY:
		return snapshot.CardArtillery, nil
	case sqlc.GameCardTypeJOLLY:
		return snapshot.CardJolly, nil
	default:
		return "", fmt.Errorf("unknown card type: %s", cardType)
	}
}
