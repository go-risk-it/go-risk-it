package gamestate

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
)

// View holds the latest game state received via WebSocket. Thread-safe.
type View struct {
	mu sync.RWMutex

	playerView *snapshot.PlayerView

	// lastUpdateTime records when the most recent Apply() was called.
	lastUpdateTime time.Time

	// version is a monotonically increasing counter incremented on each Apply().
	// Used by AwaitUpdateSince to detect updates that arrived before the wait started.
	version uint64

	// updated is closed and re-created on each state update.
	updated chan struct{}
}

func NewView() *View {
	return &View{
		updated: make(chan struct{}),
	}
}

// Version returns the current update counter. Callers snapshot this before
// triggering a server-side action, then pass it to AwaitUpdateSince to wait
// for the resulting WS update — even if it arrives before the wait starts.
func (v *View) Version() uint64 {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.version
}

// AwaitUpdateSince returns a channel that is closed when the version exceeds
// sinceVersion. If the update has already arrived (version > sinceVersion),
// returns a pre-closed channel for immediate consumption.
func (v *View) AwaitUpdateSince(sinceVersion uint64) <-chan struct{} {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.version > sinceVersion {
		ch := make(chan struct{})
		close(ch)

		return ch
	}

	return v.updated
}

// Apply processes a raw WS message and updates the relevant state.
func (v *View) Apply(msg WSMessage) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	switch msg.Type {
	case "playerView":
		var pv snapshot.PlayerView
		if err := json.Unmarshal(msg.Payload, &pv); err != nil {
			return fmt.Errorf("unmarshal playerView: %w", err)
		}

		v.playerView = &pv
	case "playerConnection":
		// Connection status updates don't affect game state — skip version bump.
		return nil
	default:
		slog.Warn("unknown ws message type", "type", msg.Type)

		return nil
	}

	v.version++
	close(v.updated)
	v.updated = make(chan struct{})
	v.lastUpdateTime = time.Now()

	return nil
}

// LastUpdateTime returns the time of the most recent Apply() call.
func (v *View) LastUpdateTime() time.Time {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.lastUpdateTime
}

// Snapshot returns a read-only copy of the current state.
func (v *View) Snapshot() ViewSnapshot {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return ViewSnapshot{
		PlayerView: v.playerView,
	}
}

// ViewSnapshot is an immutable snapshot for strategy decisions.
// The snapshot shares the pointer with the View, but View.Apply() replaces the
// pointer atomically rather than mutating the underlying struct, so concurrent
// reads of a snapshot are safe. Do not mutate snapshot fields — treat them as
// read-only.
type ViewSnapshot struct {
	PlayerView *snapshot.PlayerView
}

// MyRegions returns regions owned by the given userID.
func (s ViewSnapshot) MyRegions(userID string) []snapshot.RegionState {
	if s.PlayerView == nil {
		return nil
	}

	var regions []snapshot.RegionState
	for _, r := range s.PlayerView.Regions {
		if r.OwnerID == userID {
			regions = append(regions, r)
		}
	}

	return regions
}

// CurrentPhase returns the current phase type.
func (s ViewSnapshot) CurrentPhase() snapshot.PhaseType {
	if s.PlayerView == nil {
		return ""
	}

	return s.PlayerView.Phase.Type
}

// IsMyTurn checks if the given userID is the current player.
// Turn is a monotonically increasing counter; current player = Turn % numPlayers.
func (s ViewSnapshot) IsMyTurn(userID string) bool {
	if s.PlayerView == nil {
		return false
	}

	numPlayers := int64(len(s.PlayerView.Players))
	if numPlayers == 0 {
		return false
	}

	currentIndex := s.PlayerView.Game.Turn % numPlayers
	for _, p := range s.PlayerView.Players {
		if p.Index == currentIndex && p.UserID == userID {
			return true
		}
	}

	return false
}

// IsGameOver returns true if a winner has been declared.
func (s ViewSnapshot) IsGameOver() bool {
	return s.PlayerView != nil && s.PlayerView.Game.WinnerUserID != ""
}

// Cards returns the player's current hand of cards.
func (s ViewSnapshot) Cards() []snapshot.CardState {
	if s.PlayerView == nil {
		return nil
	}

	return s.PlayerView.Cards
}

// MyMission returns the player's assigned mission.
func (s ViewSnapshot) MyMission() snapshot.PlayerMission {
	if s.PlayerView == nil {
		return snapshot.PlayerMission{}
	}

	return s.PlayerView.Mission
}
