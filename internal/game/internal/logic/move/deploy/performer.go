package deploy

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/validation"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
)

const (
	// minTroopsToDeploy is the minimum number of troops that must be deployed in a single move.
	minTroopsToDeploy = 1
)

func (s *service) Perform(
	ctx ctx.GameContext,
	move Move,
	prev *snapshot.CachedGameState,
) (struct{}, moveservice.MoveEffect, error) {
	var zero moveservice.MoveEffect

	deployState, ok := prev.PublicSnapshot.Phase.State.(snapshot.DeployPhaseState)
	if !ok {
		return struct{}{}, zero, fmt.Errorf(
			"expected DeployPhaseState, got %T",
			prev.PublicSnapshot.Phase.State,
		)
	}

	deployableTroops := deployState.DeployableTroops

	troops := move.DesiredTroops - move.CurrentTroops
	if deployableTroops < troops {
		return struct{}{}, zero, domainerrors.NewValidationError("not enough deployable troops")
	}

	cachedRegion, err := moveservice.FindRegion(prev.PublicSnapshot.Regions, move.RegionID)
	if err != nil {
		return struct{}{}, zero, fmt.Errorf("failed to find region in cache: %w", err)
	}

	thisRegion := moveservice.ToDBRegion(cachedRegion)

	if err := s.validate(ctx, thisRegion, move, troops); err != nil {
		return struct{}{}, zero, err
	}

	observe.SpanEvent(ctx, "executing_deploy",
		attribute.String("region", thisRegion.ExternalReference),
		attribute.Int64("troops", troops),
	)

	effect := moveservice.MoveEffect{
		RegionUpdates: []moveservice.RegionUpdate{
			{
				RegionID:  thisRegion.ExternalReference,
				NewOwner:  thisRegion.UserID,
				NewTroops: thisRegion.Troops + troops,
			},
		},
		UpdatedPhase: snapshot.DeployPhaseState{
			DeployableTroops: deployableTroops - troops,
		},
		DeployableDelta: -troops,
	}

	return struct{}{}, effect, nil
}

func (s *service) validate(
	ctx ctx.GameContext,
	thisRegion *sqlc.GetRegionsByGameRow,
	move Move,
	troops int64,
) error {
	if troops < minTroopsToDeploy {
		return domainerrors.NewValidationError("must deploy at least 1 troop")
	}

	if err := validation.CheckSourceOwnedByPlayer(ctx, thisRegion, "deploy"); err != nil {
		return err
	}

	if thisRegion.Troops != move.CurrentTroops {
		observe.Warn(ctx, "declared troops mismatch",
			attribute.String("move_type", "deploy"),
			attribute.String("region", thisRegion.ExternalReference),
			attribute.Int64("actual_troops", thisRegion.Troops),
			attribute.Int64("declared_troops", move.CurrentTroops),
		)

		return domainerrors.NewValidationError(
			"region has different number of troops than declared",
		)
	}

	return nil
}
