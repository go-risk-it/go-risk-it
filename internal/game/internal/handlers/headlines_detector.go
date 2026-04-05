package handlers

import (
	"slices"
	"time"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/board"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/kernel/safego"
	"go.uber.org/fx"
)

// headlinesDetector is a stateless handler that diffs region ownership between
// the previous and current snapshots carried by MoveCompleted events. It detects
// PlayerEliminated, ContinentCaptured, and ContinentLost headlines without any
// mutable per-game state.
type headlinesDetector struct {
	continents board.Continents
	boardSvc   board.Service
	pub        eventbus.Publisher
}

// HeadlinesDetectorParams holds the dependencies for the stateless headlines detector.
type HeadlinesDetectorParams struct {
	fx.In

	Sub   eventbus.Subscriber
	Pub   eventbus.Publisher
	Board board.Service
}

// RegisterHeadlinesDetector subscribes the stateless headlines detector to
// MoveCompleted events.
func RegisterHeadlinesDetector(params HeadlinesDetectorParams) {
	det := &headlinesDetector{
		boardSvc: params.Board,
		pub:      params.Pub,
	}

	gameevt.OnGameEvent[*gameevt.MoveCompleted](params.Sub, det.handleMoveCompleted)
}

func (d *headlinesDetector) handleMoveCompleted(
	gameCtx gamectx.GameContext,
	event *gameevt.MoveCompleted,
) {
	safego.TypedSafeOp(gameCtx, "detector.headlines.v2", func(ctx gamectx.GameContext) error {
		if event.ActionType != gameapi.GamePhaseTypeATTACK {
			return nil
		}

		if event.PreviousRegions == nil || event.PublicSnapshot == nil {
			return nil
		}

		// Detect conquest via ownership diff: find a region that changed owner.
		conqueredRegion := findConqueredRegion(event.PreviousRegions, event.PublicSnapshot.Regions)
		if conqueredRegion == "" {
			return nil
		}

		if err := d.ensureContinents(ctx); err != nil {
			observe.Warn(ctx, "headlines detector: failed to load continents")

			return err
		}

		d.detectAndEmit(ctx, event, conqueredRegion)

		return nil
	})
}

func (d *headlinesDetector) ensureContinents(ctx gamectx.GameContext) error {
	if d.continents != nil {
		return nil
	}

	continents, err := d.boardSvc.GetContinents(ctx)
	if err != nil {
		return err
	}

	d.continents = continents

	return nil
}

func (d *headlinesDetector) detectAndEmit(
	gameCtx gamectx.GameContext,
	event *gameevt.MoveCompleted,
	conqueredRegion string,
) {
	previousOwnership := buildOwnershipMap(event.PreviousRegions)
	currentOwnership := buildOwnershipMap(event.PublicSnapshot.Regions)

	defenderUserID := previousOwnership[conqueredRegion]
	attackerUserID := gameCtx.UserID()

	now := time.Now()
	gameID := event.GameID()
	turn := event.Turn

	var derived []eventbus.Event

	// PlayerEliminated: defender has no regions left in the current state.
	if countRegions(currentOwnership, defenderUserID) == 0 {
		derived = append(derived, gameevt.NewPlayerEliminated(
			gameID, defenderUserID, attackerUserID, now, turn,
		))
	}

	// Continent ownership changes.
	attackerBefore := controlledContinents(d.continents, previousOwnership, attackerUserID)
	attackerAfter := controlledContinents(d.continents, currentOwnership, attackerUserID)
	defenderBefore := controlledContinents(d.continents, previousOwnership, defenderUserID)
	defenderAfter := controlledContinents(d.continents, currentOwnership, defenderUserID)

	for _, c := range attackerAfter {
		if !slices.Contains(attackerBefore, c) {
			derived = append(derived, gameevt.NewContinentCaptured(
				gameID, attackerUserID, now, c, turn,
			))
		}
	}

	for _, c := range defenderBefore {
		if !slices.Contains(defenderAfter, c) {
			derived = append(derived, gameevt.NewContinentLost(
				gameID, defenderUserID, now, c, turn,
			))
		}
	}

	for _, headline := range derived {
		d.pub.Emit(gameCtx, headline)
	}
}

// buildOwnershipMap creates a region→ownerID map from a slice of RegionState.
func buildOwnershipMap(regions []snapshot.RegionState) map[string]string {
	m := make(map[string]string, len(regions))
	for _, r := range regions {
		m[r.ID] = r.OwnerID
	}

	return m
}

// findConqueredRegion compares previous and current region ownership and returns
// the ID of the first region whose owner changed. Returns "" if no conquest occurred.
func findConqueredRegion(
	previousRegions []snapshot.RegionState,
	currentRegions []snapshot.RegionState,
) string {
	previous := buildOwnershipMap(previousRegions)
	current := buildOwnershipMap(currentRegions)

	for regionID, prevOwner := range previous {
		if curOwner, ok := current[regionID]; ok && curOwner != prevOwner {
			return regionID
		}
	}

	return ""
}

// countRegions counts how many regions a player owns in the given ownership map.
func countRegions(ownership map[string]string, userID string) int {
	count := 0
	for _, owner := range ownership {
		if owner == userID {
			count++
		}
	}

	return count
}

// controlledContinents returns the external references of all continents where
// every region is owned by the given player.
func controlledContinents(
	continents board.Continents,
	ownership map[string]string,
	userID string,
) []string {
	var controlled []string

	for _, continent := range continents.All() {
		allOwned := true

		for _, region := range continent.Regions() {
			if ownership[region] != userID {
				allOwned = false

				break
			}
		}

		if allOwned {
			controlled = append(controlled, continent.ExternalReference)
		}
	}

	return controlled
}
