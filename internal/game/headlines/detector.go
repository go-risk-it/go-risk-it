package headlines

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/snapshot"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.uber.org/fx"
)

// gameOwnership tracks per-game region ownership for incremental headline detection.
type gameOwnership struct {
	regionOwner       map[string]string // regionRef → userID
	playerRegionCount map[string]int    // userID → count of owned regions
}

// detector listens for MoveExecuted events and emits derived headline events
// (PlayerEliminated, ContinentCaptured, ContinentLost) based on ownership changes.
type detector struct {
	mu         sync.RWMutex
	games      map[int64]*gameOwnership
	continents board.Continents // lazily loaded from board service
	pub        eventbus.Publisher
	sub        eventbus.Subscriber
	snapshot   snapshot.Service
	board      board.Service
}

// DetectorParams holds the dependencies for registering the headline detector.
type DetectorParams struct {
	fx.In

	Pub      eventbus.Publisher
	Sub      eventbus.Subscriber
	Snapshot snapshot.Service
	Board    board.Service
}

// RegisterDetector subscribes the headline detector to MoveExecuted events.
func RegisterDetector(params DetectorParams) {
	det := &detector{
		games:    make(map[int64]*gameOwnership),
		pub:      params.Pub,
		sub:      params.Sub,
		snapshot: params.Snapshot,
		board:    params.Board,
	}

	gameevt.OnGameEvent[*gameevt.MoveExecuted](params.Sub, det.handleMoveExecuted)
	gameevt.OnGameEvent[*gameevt.GameCompleted](params.Sub, det.handleGameCompleted)
}

func (d *detector) handleGameCompleted(_ ctx.GameContext, event *gameevt.GameCompleted) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.games, event.GameID())
}

func (d *detector) handleMoveExecuted(gameCtx ctx.GameContext, event *gameevt.MoveExecuted) {
	_, done := observe.RawSpan(gameCtx, "detector.headlines")
	defer done(nil)

	if event.ActionType != sqlc.GamePhaseTypeATTACK {
		return
	}

	if event.AttackResult == nil || event.AttackResult.ConqueringTroops <= 0 {
		return
	}

	ownership, err := d.ensureCache(gameCtx, event.GameID())
	if err != nil {
		observe.Warn(gameCtx, "headline detector: failed to init cache")

		return
	}

	d.processConquest(gameCtx, ownership, event, gameCtx.UserID())
}

// ensureCache returns the ownership cache for the game, initializing it from a
// snapshot if this is the first event for that game.
func (d *detector) ensureCache(
	gameCtx ctx.GameContext,
	gameID int64,
) (*gameOwnership, error) {
	d.mu.RLock()
	if ownership, ok := d.games[gameID]; ok {
		d.mu.RUnlock()

		return ownership, nil
	}
	d.mu.RUnlock()

	return d.initCache(gameCtx, gameID)
}

// initCache fetches the public snapshot and builds the ownership maps. Uses a
// write lock with a double-check to avoid duplicate initialization from
// concurrent handlers. Also lazy-loads board continents on first call.
func (d *detector) initCache(
	gameCtx ctx.GameContext,
	gameID int64,
) (*gameOwnership, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Double-check after acquiring write lock
	if ownership, ok := d.games[gameID]; ok {
		return ownership, nil
	}

	if d.continents == nil {
		continents, err := d.board.GetContinents(gameCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to get continents: %w", err)
		}

		d.continents = continents
	}

	snap, err := d.snapshot.GetPublicSnapshot(gameCtx)
	if err != nil {
		return nil, err
	}

	ownership := &gameOwnership{
		regionOwner:       make(map[string]string, len(snap.Board)),
		playerRegionCount: make(map[string]int),
	}

	for _, region := range snap.Board {
		ownership.regionOwner[region.ExternalReference] = region.UserID
		ownership.playerRegionCount[region.UserID]++
	}

	d.games[gameID] = ownership

	return ownership, nil
}

// processConquest updates the ownership cache and collects headline events for a
// successful conquest (ConqueringTroops > 0). Events are emitted after releasing
// the lock to avoid deadlocks with synchronous bus implementations (TestBus).
//
// Re-emission safety: derived headline events are emitted via bus.Emit from within a
// handler goroutine. This is safe because collectHandlers uses an RLock (released before
// dispatch) and dispatchEvent launches a new goroutine per event. No lock re-entrancy.
func (d *detector) processConquest(
	eventCtx context.Context,
	ownership *gameOwnership,
	event *gameevt.MoveExecuted,
	attackerUserID string,
) {
	derived := d.detectHeadlines(ownership, event, attackerUserID)

	for _, headline := range derived {
		d.pub.Emit(eventCtx, headline)
	}
}

// ownershipDiff holds before/after continent control sets for diff-based
// headline detection.
type ownershipDiff struct {
	defenderBefore []string
	attackerBefore []string
	attackerAfter  []string
	defenderAfter  []string
}

// updateOwnership transfers a region from the defender to the attacker and returns
// the before/after continent control sets for diff-based headline detection.
func (d *detector) updateOwnership(
	ownership *gameOwnership,
	defendingRegion string,
	attackerUserID string,
) ownershipDiff {
	defenderUserID := ownership.regionOwner[defendingRegion]

	diff := ownershipDiff{
		defenderBefore: d.continentsControlledByPlayer(ownership, defenderUserID),
		attackerBefore: d.continentsControlledByPlayer(ownership, attackerUserID),
	}

	ownership.regionOwner[defendingRegion] = attackerUserID
	ownership.playerRegionCount[defenderUserID]--
	ownership.playerRegionCount[attackerUserID]++

	diff.attackerAfter = d.continentsControlledByPlayer(ownership, attackerUserID)
	diff.defenderAfter = d.continentsControlledByPlayer(ownership, defenderUserID)

	return diff
}

// detectHeadlines holds the lock, updates ownership, and returns headline events
// to be emitted outside the lock.
func (d *detector) detectHeadlines(
	ownership *gameOwnership,
	event *gameevt.MoveExecuted,
	attackerUserID string,
) []eventbus.Event {
	d.mu.Lock()
	defer d.mu.Unlock()

	defendingRegion := event.AttackResult.DefendingRegionID
	defenderUserID := ownership.regionOwner[defendingRegion]

	diff := d.updateOwnership(ownership, defendingRegion, attackerUserID)

	now := time.Now()
	gameID := event.GameID()
	turn := event.Turn

	var derived []eventbus.Event

	// PlayerEliminated if defender has no regions left
	if ownership.playerRegionCount[defenderUserID] == 0 {
		derived = append(derived, NewPlayerEliminated(
			gameID, defenderUserID, attackerUserID, now, turn,
		))
	}

	// ContinentCaptured for continents the attacker now controls but didn't before
	for _, continent := range diff.attackerAfter {
		if !slices.Contains(diff.attackerBefore, continent) {
			derived = append(derived, NewContinentCaptured(
				gameID, attackerUserID, now, continent, turn,
			))
		}
	}

	// ContinentLost for continents the defender controlled before but no longer does
	for _, continent := range diff.defenderBefore {
		if !slices.Contains(diff.defenderAfter, continent) {
			derived = append(derived, NewContinentLost(
				gameID, defenderUserID, now, continent, turn,
			))
		}
	}

	return derived
}

// continentsControlledByPlayer returns the external references of all continents
// where every region is owned by the given player.
func (d *detector) continentsControlledByPlayer(
	ownership *gameOwnership,
	userID string,
) []string {
	var controlled []string

	for _, continent := range d.continents.All() {
		if d.playerOwnsAllRegions(ownership, userID, continent) {
			controlled = append(controlled, continent.ExternalReference)
		}
	}

	return controlled
}

// playerOwnsAllRegions checks whether the given player owns every region in the
// continent according to the current ownership cache.
func (d *detector) playerOwnsAllRegions(
	ownership *gameOwnership,
	userID string,
	continent *board.Continent,
) bool {
	for _, region := range continent.Regions() {
		if ownership.regionOwner[region] != userID {
			return false
		}
	}

	return true
}
