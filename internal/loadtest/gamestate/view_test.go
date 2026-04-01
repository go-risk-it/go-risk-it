package gamestate_test

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeMsg(t *testing.T, msgType string, payload any) gamestate.WSMessage {
	t.Helper()

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	return gamestate.WSMessage{Type: msgType, Payload: data}
}

func TestView_ApplyGameState(t *testing.T) {
	t.Parallel()

	v := gamestate.NewView()

	gs := gamestate.GameState{
		ID:   1,
		Turn: 5,
		Phase: gamestate.Phase{
			Type: gamestate.Deploy,
		},
	}

	err := v.Apply(makeMsg(t, "gameState", gs))
	require.NoError(t, err)

	snap := v.Snapshot()
	assert.Equal(t, int64(1), snap.GameState.ID)
	assert.Equal(t, int64(5), snap.GameState.Turn)
	assert.Equal(t, gamestate.Deploy, snap.GameState.Phase.Type)
}

func TestView_ApplyBoardState(t *testing.T) {
	t.Parallel()

	v := gamestate.NewView()

	bs := gamestate.BoardState{
		Regions: []gamestate.Region{
			{ID: "r1", OwnerID: "user1", Troops: 3},
			{ID: "r2", OwnerID: "user2", Troops: 5},
		},
	}

	err := v.Apply(makeMsg(t, "boardState", bs))
	require.NoError(t, err)

	snap := v.Snapshot()
	assert.Len(t, snap.BoardState.Regions, 2)
	assert.Equal(t, "r1", snap.BoardState.Regions[0].ID)
}

func TestView_ApplyPlayerState(t *testing.T) {
	t.Parallel()

	v := gamestate.NewView()

	ps := gamestate.PlayersState{
		Players: []gamestate.Player{
			{UserID: "u1", Index: 0},
			{UserID: "u2", Index: 1},
		},
	}

	err := v.Apply(makeMsg(t, "playerState", ps))
	require.NoError(t, err)

	snap := v.Snapshot()
	assert.Len(t, snap.PlayersState.Players, 2)
}

func TestView_ApplyCardState(t *testing.T) {
	t.Parallel()

	v := gamestate.NewView()

	cs := gamestate.CardState{
		Cards: []gamestate.Card{
			{ID: 1, Type: gamestate.Infantry, Region: "r1"},
		},
	}

	err := v.Apply(makeMsg(t, "cardState", cs))
	require.NoError(t, err)

	snap := v.Snapshot()
	assert.Len(t, snap.CardState.Cards, 1)
}

func TestView_ApplyIgnoredTypes(t *testing.T) {
	t.Parallel()

	v := gamestate.NewView()

	for _, msgType := range []string{"moveHistory", "missionState", "lobbyState"} {
		err := v.Apply(gamestate.WSMessage{Type: msgType, Payload: json.RawMessage(`{}`)})
		assert.NoError(t, err, "should ignore %s", msgType)
	}
}

func TestView_ApplyUnknownType(t *testing.T) {
	t.Parallel()

	v := gamestate.NewView()

	err := v.Apply(gamestate.WSMessage{Type: "bogus", Payload: json.RawMessage(`{}`)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown message type")
}

func TestView_ApplyInvalidJSON(t *testing.T) {
	t.Parallel()

	v := gamestate.NewView()

	err := v.Apply(gamestate.WSMessage{Type: "gameState", Payload: json.RawMessage(`{bad json`)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal gameState")
}

func TestView_UpdatedChannel(t *testing.T) {
	t.Parallel()

	v := gamestate.NewView()

	ch := v.Updated()

	// Channel should be open.
	select {
	case <-ch:
		t.Fatal("channel should not be closed yet")
	default:
	}

	// Apply triggers close.
	err := v.Apply(gamestate.WSMessage{Type: "lobbyState", Payload: json.RawMessage(`{}`)})
	require.NoError(t, err)

	select {
	case <-ch:
		// OK — channel was closed.
	default:
		t.Fatal("channel should be closed after Apply")
	}

	// New channel should be open.
	ch2 := v.Updated()
	select {
	case <-ch2:
		t.Fatal("new channel should not be closed")
	default:
	}
}

func TestView_LastUpdateTime(t *testing.T) {
	t.Parallel()

	v := gamestate.NewView()
	assert.True(t, v.LastUpdateTime().IsZero())

	err := v.Apply(gamestate.WSMessage{Type: "lobbyState", Payload: json.RawMessage(`{}`)})
	require.NoError(t, err)

	assert.False(t, v.LastUpdateTime().IsZero())
}

func TestViewSnapshot_MyRegions(t *testing.T) {
	t.Parallel()

	snap := gamestate.ViewSnapshot{
		BoardState: &gamestate.BoardState{
			Regions: []gamestate.Region{
				{ID: "r1", OwnerID: "alice"},
				{ID: "r2", OwnerID: "bob"},
				{ID: "r3", OwnerID: "alice"},
			},
		},
	}

	assert.Len(t, snap.MyRegions("alice"), 2)
	assert.Len(t, snap.MyRegions("bob"), 1)
	assert.Nil(t, snap.MyRegions("charlie"))
}

func TestViewSnapshot_MyRegions_NilBoard(t *testing.T) {
	t.Parallel()

	snap := gamestate.ViewSnapshot{}
	assert.Nil(t, snap.MyRegions("alice"))
}

func TestViewSnapshot_CurrentPhase(t *testing.T) {
	t.Parallel()

	snap := gamestate.ViewSnapshot{
		GameState: &gamestate.GameState{Phase: gamestate.Phase{Type: gamestate.Attack}},
	}
	assert.Equal(t, gamestate.Attack, snap.CurrentPhase())
}

func TestViewSnapshot_CurrentPhase_NilGame(t *testing.T) {
	t.Parallel()

	snap := gamestate.ViewSnapshot{}
	assert.Equal(t, gamestate.PhaseType(""), snap.CurrentPhase())
}

func TestViewSnapshot_IsMyTurn(t *testing.T) {
	t.Parallel()

	snap := gamestate.ViewSnapshot{
		GameState: &gamestate.GameState{Turn: 3},
		PlayersState: &gamestate.PlayersState{
			Players: []gamestate.Player{
				{UserID: "alice", Index: 0},
				{UserID: "bob", Index: 1},
				{UserID: "charlie", Index: 2},
				{UserID: "diana", Index: 3},
			},
		},
	}

	// Turn 3 % 4 players = index 3 → diana.
	assert.True(t, snap.IsMyTurn("diana"))
	assert.False(t, snap.IsMyTurn("alice"))
	assert.False(t, snap.IsMyTurn("bob"))
}

func TestViewSnapshot_IsMyTurn_EdgeCases(t *testing.T) {
	t.Parallel()

	// Nil game state.
	snap := gamestate.ViewSnapshot{}
	assert.False(t, snap.IsMyTurn("anyone"))

	// No players.
	snap = gamestate.ViewSnapshot{
		GameState:    &gamestate.GameState{Turn: 1},
		PlayersState: &gamestate.PlayersState{},
	}
	assert.False(t, snap.IsMyTurn("anyone"))
}

func TestViewSnapshot_IsGameOver(t *testing.T) {
	t.Parallel()

	snap := gamestate.ViewSnapshot{
		GameState: &gamestate.GameState{WinnerUserID: "alice"},
	}
	assert.True(t, snap.IsGameOver())

	snap = gamestate.ViewSnapshot{
		GameState: &gamestate.GameState{},
	}
	assert.False(t, snap.IsGameOver())

	snap = gamestate.ViewSnapshot{}
	assert.False(t, snap.IsGameOver())
}

func TestView_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	v := gamestate.NewView()

	var wg sync.WaitGroup

	// Concurrent writers.
	for i := range 10 {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			gs := gamestate.GameState{Turn: int64(idx)}
			data, _ := json.Marshal(gs)
			_ = v.Apply(gamestate.WSMessage{Type: "gameState", Payload: data})
		}(i)
	}

	// Concurrent readers.
	for range 10 {
		wg.Go(func() {
			_ = v.Snapshot()
			_ = v.Updated()
			_ = v.LastUpdateTime()
		})
	}

	wg.Wait()
}
