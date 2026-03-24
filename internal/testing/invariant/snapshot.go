//go:build invariant

package invariant

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
)

// GameSnapshot captures the full game state at a point in time.
// All invariant checkers receive snapshots, not raw DB handles.
type GameSnapshot struct {
	GameID  int64
	Phase   sqlc.GamePhaseType
	Turn    int64
	Winner  string
	Regions []sqlc.GetRegionsByGameRow
	Players []sqlc.GetPlayersStateRow
}

// TakeSnapshot queries the database and builds a GameSnapshot.
func TakeSnapshot(
	tb testing.TB,
	harness *Harness,
	gameID int64,
) *GameSnapshot {
	tb.Helper()

	gCtx := harness.GameCtx(gameID, "snapshot-reader")

	gameState, err := harness.StateService.GetGameState(gCtx)
	if err != nil {
		tb.Fatalf("failed to get game state: %v", err)
	}

	regions, err := harness.Querier.GetRegionsByGame(gCtx, gameID)
	if err != nil {
		tb.Fatalf("failed to get regions: %v", err)
	}

	players, err := harness.Querier.GetPlayersState(gCtx, gameID)
	if err != nil {
		tb.Fatalf("failed to get players state: %v", err)
	}

	return &GameSnapshot{
		GameID:  gameID,
		Phase:   gameState.Phase,
		Turn:    gameState.Turn,
		Winner:  gameState.WinnerUserID,
		Regions: regions,
		Players: players,
	}
}

// CurrentPlayerUserID derives the current player from turn index.
func (s *GameSnapshot) CurrentPlayerUserID() string {
	turnIndex := s.Turn % int64(len(s.Players))

	for _, player := range s.Players {
		if player.TurnIndex == turnIndex {
			return player.UserID
		}
	}

	return ""
}

// RegionsOwnedBy returns all regions owned by the given user.
func (s *GameSnapshot) RegionsOwnedBy(
	userID string,
) []sqlc.GetRegionsByGameRow {
	var result []sqlc.GetRegionsByGameRow
	for _, region := range s.Regions {
		if region.UserID == userID {
			result = append(result, region)
		}
	}

	return result
}

// TotalTroops returns the sum of all troops across all regions.
func (s *GameSnapshot) TotalTroops() int64 {
	var total int64
	for _, region := range s.Regions {
		total += region.Troops
	}

	return total
}

// IsGameOver returns true if the game has a winner.
func (s *GameSnapshot) IsGameOver() bool {
	return s.Winner != ""
}

// ActivePlayerCount returns the number of players who still own
// at least one region.
func (s *GameSnapshot) ActivePlayerCount() int {
	count := 0
	for _, player := range s.Players {
		if player.RegionCount > 0 {
			count++
		}
	}

	return count
}
