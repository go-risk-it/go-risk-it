package orchestration

import (
	"fmt"

	apisnapshot "github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/phase"
	kerneldata "github.com/go-risk-it/go-risk-it/internal/kernel/data"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Persist executes all database writes in a PersistenceEffect within a single
// ReadCommitted transaction. Writes are grouped into 6 categories executed in
// deterministic order:
//
//  1. MoveLog — move history (references game + player)
//  2. MoveExecution — region mutations (troops, ownership, deployable, card unlinks, bonuses)
//  3. Elimination — cascade data (transfer cards, reassign missions, delete spurious)
//  4. CardDraw — independent card ownership changes
//  5. PhaseTransition — creates new FK-referenced rows (phase, conquer_phase, deploy_phase)
//  6. GameConclusion — updates game row (winner)
//
// Nil groups are skipped. A span event is emitted for each non-nil group.
// Errors from any write propagate immediately (no partial success).
//
// The phaseService dependency is needed for InsertPhase calls in PhaseTransition.
func Persist(
	gameCtx ctx.GameContext,
	querier db.Querier,
	phaseService phase.Service,
	effect *PersistenceEffect,
) error {
	_, err := kerneldata.InTransactionWithIsolation(
		querier,
		gameCtx,
		nil, // no metrics
		pgx.ReadCommitted,
		func(txQuerier db.Querier) (struct{}, error) {
			if err := persistGroups(
				gameCtx, txQuerier, phaseService, effect,
			); err != nil {
				return struct{}{}, err
			}

			return struct{}{}, nil
		},
	)

	return err
}

// persistStep represents a single named persistence operation that can be
// skipped (when active is false) or executed (when active is true).
type persistStep struct {
	active    bool
	spanEvent string
	execute   func() error
}

// persistGroups executes each non-nil effect group in FK-dependency order.
// Uses a table-driven approach to keep cyclomatic complexity low.
func persistGroups(
	gameCtx ctx.GameContext,
	txQuerier db.Querier,
	phaseService phase.Service,
	effect *PersistenceEffect,
) error {
	steps := buildPersistSteps(gameCtx, txQuerier, phaseService, effect)

	for _, step := range steps {
		if !step.active {
			continue
		}

		observe.SpanEvent(gameCtx, step.spanEvent)

		if err := step.execute(); err != nil {
			return err
		}
	}

	return nil
}

// buildPersistSteps constructs the ordered list of persistence steps from
// the effect. Each step's active flag reflects whether the group is non-nil.
func buildPersistSteps(
	gameCtx ctx.GameContext,
	txQuerier db.Querier,
	phaseService phase.Service,
	effect *PersistenceEffect,
) []persistStep {
	return []persistStep{
		{
			active:    effect.MoveLog != nil,
			spanEvent: "persist.move_log",
			execute: func() error {
				return persistMoveLog(gameCtx, txQuerier, effect.MoveLog)
			},
		},
		{
			active:    effect.MoveExecution != nil,
			spanEvent: "persist.move_execution",
			execute: func() error {
				return persistMoveExecution(
					gameCtx, txQuerier, effect.MoveExecution,
				)
			},
		},
		{
			active:    effect.Elimination != nil,
			spanEvent: "persist.elimination",
			execute: func() error {
				return persistElimination(
					gameCtx, txQuerier, effect.Elimination,
				)
			},
		},
		{
			active:    effect.CardDraw != nil,
			spanEvent: "persist.card_draw",
			execute: func() error {
				return persistCardDraw(gameCtx, txQuerier, effect.CardDraw)
			},
		},
		{
			active:    effect.PhaseTransition != nil,
			spanEvent: "persist.phase_transition",
			execute: func() error {
				return persistPhaseTransition(
					gameCtx,
					txQuerier,
					phaseService,
					effect.PhaseTransition,
				)
			},
		},
		{
			active:    effect.GameConclusion != nil,
			spanEvent: "persist.game_conclusion",
			execute: func() error {
				return persistGameConclusion(
					gameCtx, txQuerier, effect.GameConclusion,
				)
			},
		},
	}
}

// persistMoveLog writes the move log entry. CreateMoveLog resolves the UserID
// to an internal player ID via subquery.
func persistMoveLog(
	gameCtx ctx.GameContext,
	querier db.Querier,
	moveLog *MoveLogEntry,
) error {
	_, err := querier.CreateMoveLog(gameCtx, sqlc.CreateMoveLogParams{
		GameID:   moveLog.GameID,
		UserID:   moveLog.UserID,
		Phase:    sqlc.GamePhaseType(moveLog.PhaseType),
		MoveData: moveLog.MoveData,
		Result:   moveLog.Result,
	})
	if err != nil {
		return fmt.Errorf("CreateMoveLog failed: %w", err)
	}

	return nil
}

// persistMoveExecution writes all region-level and player-level mutations in order:
//  1. RegionTroopUpdates (signed deltas applied to region troops)
//  2. OwnershipChanges (region conquest)
//  3. DeployableDelta (player deployable troops change)
//  4. CardUnlinks (card IDs unlinked from owner)
//  5. RegionBonuses (troops granted to multiple regions)
func persistMoveExecution(
	gameCtx ctx.GameContext,
	querier db.Querier,
	execution *MoveExecution,
) error {
	if err := persistRegionTroopUpdates(gameCtx, querier, execution); err != nil {
		return err
	}

	if err := persistOwnershipChanges(gameCtx, querier, execution); err != nil {
		return err
	}

	if err := persistDeployableDelta(gameCtx, querier, execution); err != nil {
		return err
	}

	if err := persistCardUnlinks(gameCtx, querier, execution); err != nil {
		return err
	}

	return persistRegionBonuses(gameCtx, querier, execution)
}

// persistRegionTroopUpdates applies signed troop deltas to regions.
func persistRegionTroopUpdates(
	gameCtx ctx.GameContext,
	querier db.Querier,
	execution *MoveExecution,
) error {
	for _, update := range execution.RegionTroopUpdates {
		err := querier.IncreaseRegionTroops(gameCtx, sqlc.IncreaseRegionTroopsParams{
			ID:     update.RegionID,
			Troops: update.Delta,
		})
		if err != nil {
			return fmt.Errorf(
				"IncreaseRegionTroops failed for region %d: %w",
				update.RegionID,
				err,
			)
		}
	}

	return nil
}

// persistOwnershipChanges writes region conquest ownership transfers.
func persistOwnershipChanges(
	gameCtx ctx.GameContext,
	querier db.Querier,
	execution *MoveExecution,
) error {
	for _, change := range execution.OwnershipChanges {
		_, err := querier.UpdateRegionOwner(gameCtx, sqlc.UpdateRegionOwnerParams{
			GameID:            gameCtx.GameID(),
			NewOwnerUserID:    change.NewOwnerUserID,
			ConqueredRegionID: change.RegionID,
		})
		if err != nil {
			return fmt.Errorf(
				"UpdateRegionOwner failed for region %d: %w",
				change.RegionID,
				err,
			)
		}
	}

	return nil
}

// persistDeployableDelta decreases a player's deployable troop count.
func persistDeployableDelta(
	gameCtx ctx.GameContext,
	querier db.Querier,
	execution *MoveExecution,
) error {
	if execution.DeployableDelta == nil {
		return nil
	}

	absoluteDelta := execution.DeployableDelta.Delta
	if absoluteDelta < 0 {
		absoluteDelta = -absoluteDelta
	}

	err := querier.DecreaseDeployableTroops(
		gameCtx,
		sqlc.DecreaseDeployableTroopsParams{
			ID:               execution.DeployableDelta.PlayerID,
			DeployableTroops: absoluteDelta,
		},
	)
	if err != nil {
		return fmt.Errorf("DecreaseDeployableTroops failed: %w", err)
	}

	return nil
}

// persistCardUnlinks removes card-owner links for played cards.
func persistCardUnlinks(
	gameCtx ctx.GameContext,
	querier db.Querier,
	execution *MoveExecution,
) error {
	if len(execution.CardUnlinks) == 0 {
		return nil
	}

	err := querier.UnlinkCardsFromOwner(gameCtx, execution.CardUnlinks)
	if err != nil {
		return fmt.Errorf("UnlinkCardsFromOwner failed: %w", err)
	}

	return nil
}

// persistRegionBonuses grants troops to multiple regions simultaneously.
func persistRegionBonuses(
	gameCtx ctx.GameContext,
	querier db.Querier,
	execution *MoveExecution,
) error {
	if execution.RegionBonuses == nil {
		return nil
	}

	err := querier.GrantRegionTroops(gameCtx, sqlc.GrantRegionTroopsParams{
		Troops:  execution.RegionBonuses.TroopsPerRegion,
		Regions: execution.RegionBonuses.RegionIDs,
	})
	if err != nil {
		return fmt.Errorf("GrantRegionTroops failed: %w", err)
	}

	return nil
}

// persistElimination executes the 3-write cascade triggered by player elimination:
//  1. Resolve EliminatedUserID to internal player ID via GetPlayerByUserId
//  2. TransferCardsOwnership (eliminated player's cards → conqueror)
//  3. ReassignMissions (missions targeting eliminated player → TwentyFourTerritories)
//  4. DeleteSpuriousEliminatePlayerMissions (missions targeting eliminated player)
func persistElimination(
	gameCtx ctx.GameContext,
	querier db.Querier,
	elimination *EliminationEffect,
) error {
	// 1. Resolve eliminated player's internal ID
	eliminatedPlayer, err := querier.GetPlayerByUserId(gameCtx, elimination.EliminatedUserID)
	if err != nil {
		return fmt.Errorf(
			"GetPlayerByUserId failed for eliminated player %s: %w",
			elimination.EliminatedUserID,
			err,
		)
	}

	// 2. Transfer cards
	err = querier.TransferCardsOwnership(gameCtx, sqlc.TransferCardsOwnershipParams{
		GameID: gameCtx.GameID(),
		To:     elimination.ConquerorUserID,
		From:   pgtype.Int8{Int64: eliminatedPlayer.ID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("TransferCardsOwnership failed: %w", err)
	}

	// 3. Reassign missions
	err = querier.ReassignMissions(gameCtx, sqlc.ReassignMissionsParams{
		GameID:             gameCtx.GameID(),
		UserID:             elimination.EliminatedUserID,
		EliminatedPlayerID: eliminatedPlayer.ID,
	})
	if err != nil {
		return fmt.Errorf("ReassignMissions failed: %w", err)
	}

	// 4. Delete spurious eliminate-player missions
	err = querier.DeleteSpuriousEliminatePlayerMissions(gameCtx, gameCtx.GameID())
	if err != nil {
		return fmt.Errorf("DeleteSpuriousEliminatePlayerMissions failed: %w", err)
	}

	return nil
}

// persistCardDraw writes a single card draw. DrawCard resolves the UserID to
// an internal player ID via subquery.
func persistCardDraw(
	gameCtx ctx.GameContext,
	querier db.Querier,
	cardDraw *CardDraw,
) error {
	err := querier.DrawCard(gameCtx, sqlc.DrawCardParams{
		ID:     cardDraw.CardID,
		UserID: cardDraw.UserID,
		GameID: gameCtx.GameID(),
	})
	if err != nil {
		return fmt.Errorf("DrawCard failed: %w", err)
	}

	return nil
}

// persistPhaseTransition creates a new phase row and optional sub-phase rows:
//  1. InsertPhase (via phase.Service) → phase ID
//  2. InsertConquerPhase or InsertDeployPhase (sequential dependency on phase ID)
func persistPhaseTransition(
	gameCtx ctx.GameContext,
	querier db.Querier,
	phaseService phase.Service,
	transition *PhaseTransition,
) error {
	// Convert PlayerRef slice to snapshot.PlayerState slice
	players := make([]apisnapshot.PlayerState, len(transition.Players))
	for i, p := range transition.Players {
		players[i] = apisnapshot.PlayerState{UserID: p.UserID}
	}

	// 1. InsertPhase
	phaseRecord, err := phaseService.InsertPhase(gameCtx, querier, phase.PhaseInsertParams{
		PhaseType:    sqlc.GamePhaseType(transition.PhaseType),
		CurrentPhase: sqlc.GamePhaseType(transition.CurrentPhaseType),
		Turn:         transition.Turn,
		Players:      players,
	})
	if err != nil {
		return fmt.Errorf("InsertPhase failed: %w", err)
	}

	// 2. InsertConquerPhase or InsertDeployPhase
	if transition.ConquerData != nil {
		_, err := querier.InsertConquerPhase(gameCtx, sqlc.InsertConquerPhaseParams{
			PhaseID:             phaseRecord.ID,
			ID:                  gameCtx.GameID(),
			ExternalReference:   transition.ConquerData.SourceRegionName,
			ExternalReference_2: transition.ConquerData.TargetRegionName,
			MinimumTroops:       transition.ConquerData.MinTroops,
		})
		if err != nil {
			return fmt.Errorf("InsertConquerPhase failed: %w", err)
		}
	} else if transition.DeployData != nil {
		_, err := querier.InsertDeployPhase(gameCtx, sqlc.InsertDeployPhaseParams{
			PhaseID:          phaseRecord.ID,
			DeployableTroops: transition.DeployData.DeployableTroops,
		})
		if err != nil {
			return fmt.Errorf("InsertDeployPhase failed: %w", err)
		}
	}

	return nil
}

// persistGameConclusion assigns the winner to the game. Resolves WinnerUserID
// to an internal player ID via GetPlayerByUserId.
func persistGameConclusion(
	gameCtx ctx.GameContext,
	querier db.Querier,
	conclusion *GameConclusion,
) error {
	// 1. Resolve winner's internal ID
	winnerPlayer, err := querier.GetPlayerByUserId(gameCtx, conclusion.WinnerUserID)
	if err != nil {
		return fmt.Errorf(
			"GetPlayerByUserId failed for winner %s: %w",
			conclusion.WinnerUserID,
			err,
		)
	}

	// 2. Assign game winner
	err = querier.AssignGameWinner(gameCtx, sqlc.AssignGameWinnerParams{
		WinnerPlayerID: pgtype.Int8{Int64: winnerPlayer.ID, Valid: true},
		GameID:         gameCtx.GameID(),
	})
	if err != nil {
		return fmt.Errorf("AssignGameWinner failed: %w", err)
	}

	return nil
}
