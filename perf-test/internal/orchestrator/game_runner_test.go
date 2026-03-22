package orchestrator

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/gamestate"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player"
)

func TestFindActivePlayer(t *testing.T) {
	t.Parallel()

	mkPlayers := func(ids ...string) []*PlayerInfo {
		ps := make([]*PlayerInfo, len(ids))
		for i, id := range ids {
			ps[i] = &PlayerInfo{UserID: id, Name: "p" + id}
		}

		return ps
	}

	mkIndex := func(players []*PlayerInfo) map[string]int {
		idx := make(map[string]int)
		for i, p := range players {
			idx[p.UserID] = i
		}

		return idx
	}

	mkPlayersState := func(entries ...struct {
		userID string
		index  int64
	},
	) *gamestate.PlayersState {
		ps := &gamestate.PlayersState{Players: make([]gamestate.Player, len(entries))}
		for i, e := range entries {
			ps.Players[i] = gamestate.Player{UserID: e.userID, Index: e.index}
		}

		return ps
	}

	type entry struct {
		userID string
		index  int64
	}

	players3 := mkPlayers("u0", "u1", "u2")
	index3 := mkIndex(players3)
	ps3 := mkPlayersState(
		entry{"u0", 0},
		entry{"u1", 1},
		entry{"u2", 2},
	)

	tests := []struct {
		name      string
		snap      gamestate.ViewSnapshot
		players   []*PlayerInfo
		userIndex map[string]int
		wantIdx   int
		wantNil   bool
	}{
		{
			name: "turn 0 with 3 players returns player 0",
			snap: gamestate.ViewSnapshot{
				GameState:    &gamestate.GameState{Turn: 0},
				PlayersState: ps3,
			},
			players:   players3,
			userIndex: index3,
			wantIdx:   0,
		},
		{
			name: "turn 1 with 3 players returns player 1",
			snap: gamestate.ViewSnapshot{
				GameState:    &gamestate.GameState{Turn: 1},
				PlayersState: ps3,
			},
			players:   players3,
			userIndex: index3,
			wantIdx:   1,
		},
		{
			name: "turn 2 with 3 players returns player 2",
			snap: gamestate.ViewSnapshot{
				GameState:    &gamestate.GameState{Turn: 2},
				PlayersState: ps3,
			},
			players:   players3,
			userIndex: index3,
			wantIdx:   2,
		},
		{
			name: "turn wraps around: turn 3 with 3 players returns player 0",
			snap: gamestate.ViewSnapshot{
				GameState:    &gamestate.GameState{Turn: 3},
				PlayersState: ps3,
			},
			players:   players3,
			userIndex: index3,
			wantIdx:   0,
		},
		{
			name: "turn wraps around: turn 7 with 3 players returns player 1",
			snap: gamestate.ViewSnapshot{
				GameState:    &gamestate.GameState{Turn: 7},
				PlayersState: ps3,
			},
			players:   players3,
			userIndex: index3,
			wantIdx:   1,
		},
		{
			name: "nil gameState returns nil",
			snap: gamestate.ViewSnapshot{
				GameState:    nil,
				PlayersState: ps3,
			},
			players:   players3,
			userIndex: index3,
			wantNil:   true,
		},
		{
			name: "nil playersState returns nil",
			snap: gamestate.ViewSnapshot{
				GameState:    &gamestate.GameState{Turn: 0},
				PlayersState: nil,
			},
			players:   players3,
			userIndex: index3,
			wantNil:   true,
		},
		{
			name: "empty players returns nil",
			snap: gamestate.ViewSnapshot{
				GameState:    &gamestate.GameState{Turn: 0},
				PlayersState: &gamestate.PlayersState{Players: []gamestate.Player{}},
			},
			players:   players3,
			userIndex: index3,
			wantNil:   true,
		},
		{
			name: "player not found in userIndex returns nil",
			snap: gamestate.ViewSnapshot{
				GameState: &gamestate.GameState{Turn: 0},
				PlayersState: &gamestate.PlayersState{
					Players: []gamestate.Player{
						{UserID: "unknown-user", Index: 0},
					},
				},
			},
			players:   players3,
			userIndex: index3,
			wantNil:   true,
		},
		{
			name: "both nil returns nil",
			snap: gamestate.ViewSnapshot{
				GameState:    nil,
				PlayersState: nil,
			},
			players:   players3,
			userIndex: index3,
			wantNil:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			idx, p := findActivePlayer(tc.snap, tc.players, tc.userIndex)
			if tc.wantNil {
				if p != nil {
					t.Fatalf("expected nil player, got %+v (idx=%d)", p, idx)
				}

				if idx != -1 {
					t.Fatalf("expected idx -1 for nil player, got %d", idx)
				}

				return
			}

			if p == nil {
				t.Fatal("expected non-nil player, got nil")
			}

			if idx != tc.wantIdx {
				t.Fatalf("expected idx %d, got %d", tc.wantIdx, idx)
			}

			if p != tc.players[tc.wantIdx] {
				t.Fatalf("expected player pointer at index %d, got different pointer", tc.wantIdx)
			}
		})
	}
}

func TestActionTypeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		action player.ActionType
		want   string
	}{
		{player.ActionDeploy, "deploy"},
		{player.ActionAttack, "attack"},
		{player.ActionConquer, "conquer"},
		{player.ActionReinforce, "reinforce"},
		{player.ActionPlayCards, "cards"},
		{player.ActionAdvance, "advance"},
		{player.ActionType(999), "unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()

			got := actionTypeName(tc.action)
			if got != tc.want {
				t.Fatalf("actionTypeName(%d) = %q, want %q", tc.action, got, tc.want)
			}
		})
	}
}

func TestAbbreviate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"long string truncated to 8", "abcdefghijklmnop", "abcdefgh"},
		{"exactly 8 chars unchanged", "abcdefgh", "abcdefgh"},
		{"exactly 9 chars truncated", "abcdefghi", "abcdefgh"},
		{"short string unchanged", "abc", "abc"},
		{"empty string", "", ""},
		{"single char", "x", "x"},
		{"short with trailing space trimmed", "abc ", "abc"},
		{"short with leading space trimmed", " abc", "abc"},
		{"long string ignores spaces in middle", "ab cd ef gh ij", "ab cd ef"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := abbreviate(tc.input)
			if got != tc.want {
				t.Fatalf("abbreviate(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
