package gamestate

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// View holds the latest game state received via WebSocket. Thread-safe.
type View struct {
	mu sync.RWMutex

	GameState    *GameState
	BoardState   *BoardState
	PlayersState *PlayersState
	CardState    *CardState

	// lastUpdateTime records when the most recent Apply() was called.
	lastUpdateTime time.Time

	// updated is closed and re-created on each state update.
	updated chan struct{}
}

func NewView() *View {
	return &View{
		updated: make(chan struct{}),
	}
}

// Updated returns a channel that is closed when a new state update arrives.
func (v *View) Updated() <-chan struct{} {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.updated
}

// Apply processes a raw WS message and updates the relevant state.
func (v *View) Apply(msg WSMessage) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	switch msg.Type {
	case "gameState":
		var gs GameState
		if err := json.Unmarshal(msg.Payload, &gs); err != nil {
			return fmt.Errorf("unmarshal gameState: %w", err)
		}

		v.GameState = &gs
	case "boardState":
		var bs BoardState
		if err := json.Unmarshal(msg.Payload, &bs); err != nil {
			return fmt.Errorf("unmarshal boardState: %w", err)
		}

		v.BoardState = &bs
	case "playerState":
		var ps PlayersState
		if err := json.Unmarshal(msg.Payload, &ps); err != nil {
			return fmt.Errorf("unmarshal playerState: %w", err)
		}

		v.PlayersState = &ps
	case "cardState":
		var cs CardState
		if err := json.Unmarshal(msg.Payload, &cs); err != nil {
			return fmt.Errorf("unmarshal cardState: %w", err)
		}

		v.CardState = &cs
	case "moveHistory", "missionState", "lobbyState":
		// Ignored for perf testing
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}

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
		GameState:    v.GameState,
		BoardState:   v.BoardState,
		PlayersState: v.PlayersState,
		CardState:    v.CardState,
	}
}

// ViewSnapshot is an immutable snapshot for strategy decisions.
// The snapshot shares pointer values with the View (e.g., *GameState), but
// View.Apply() replaces pointers atomically rather than mutating the underlying
// structs, so concurrent reads of a snapshot are safe. Do not mutate snapshot
// fields — treat them as read-only.
type ViewSnapshot struct {
	GameState    *GameState
	BoardState   *BoardState
	PlayersState *PlayersState
	CardState    *CardState
}

// MyRegions returns regions owned by the given userID.
func (s ViewSnapshot) MyRegions(userID string) []Region {
	if s.BoardState == nil {
		return nil
	}

	var regions []Region
	for _, r := range s.BoardState.Regions {
		if r.OwnerID == userID {
			regions = append(regions, r)
		}
	}

	return regions
}

// CurrentPhase returns the current phase type.
func (s ViewSnapshot) CurrentPhase() PhaseType {
	if s.GameState == nil {
		return ""
	}

	return s.GameState.Phase.Type
}

// IsMyTurn checks if the given userID is the current player.
// Turn is a monotonically increasing counter; current player = Turn % numPlayers.
func (s ViewSnapshot) IsMyTurn(userID string) bool {
	if s.GameState == nil || s.PlayersState == nil {
		return false
	}

	numPlayers := int64(len(s.PlayersState.Players))
	if numPlayers == 0 {
		return false
	}

	currentIndex := s.GameState.Turn % numPlayers
	for _, p := range s.PlayersState.Players {
		if p.Index == currentIndex && p.UserID == userID {
			return true
		}
	}

	return false
}

// IsGameOver returns true if a winner has been declared.
func (s ViewSnapshot) IsGameOver() bool {
	return s.GameState != nil && s.GameState.WinnerUserID != ""
}
