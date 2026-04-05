package gamestate_test

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testPlayerView returns a minimal but complete PlayerView for testing.
func testPlayerView() *snapshot.PlayerView {
	return &snapshot.PlayerView{
		Game: snapshot.GameMeta{
			ID:   1,
			Turn: 5,
		},
		Phase: snapshot.Phase{
			Type:  snapshot.PhaseDeploy,
			State: snapshot.DeployPhaseState{DeployableTroops: 3},
		},
		Regions: []snapshot.RegionState{
			{ID: "r1", OwnerID: "alice", Troops: 3},
			{ID: "r2", OwnerID: "bob", Troops: 5},
			{ID: "r3", OwnerID: "alice", Troops: 1},
		},
		Players: []snapshot.PlayerState{
			{UserID: "alice", Name: "Alice", Index: 0, CardCount: 2, Status: snapshot.PlayerAlive},
			{UserID: "bob", Name: "Bob", Index: 1, CardCount: 1, Status: snapshot.PlayerAlive},
		},
		Cards: []snapshot.CardState{
			{ID: 1, Type: snapshot.CardInfantry, Region: "r1"},
			{ID: 2, Type: snapshot.CardCavalry, Region: "r2"},
		},
		Mission: snapshot.PlayerMission{
			Type:   snapshot.MissionTwentyFourTerritories,
			Detail: snapshot.TwentyFourTerritoriesMission{},
		},
	}
}

func makePlayerViewMsg(t *testing.T, pv *snapshot.PlayerView) gamestate.WSMessage {
	t.Helper()

	data, err := json.Marshal(pv)
	require.NoError(t, err)

	return gamestate.WSMessage{Type: "playerView", Payload: data}
}

func TestView_ApplyPlayerView(t *testing.T) {
	t.Parallel()

	v := gamestate.NewView()
	pv := testPlayerView()

	err := v.Apply(makePlayerViewMsg(t, pv))
	require.NoError(t, err)

	snap := v.Snapshot()
	require.NotNil(t, snap.PlayerView)
	assert.Equal(t, int64(1), snap.PlayerView.Game.ID)
	assert.Equal(t, int64(5), snap.PlayerView.Game.Turn)
	assert.Equal(t, snapshot.PhaseDeploy, snap.PlayerView.Phase.Type)
	assert.Len(t, snap.PlayerView.Regions, 3)
	assert.Len(t, snap.PlayerView.Players, 2)
	assert.Len(t, snap.PlayerView.Cards, 2)
}

func TestView_ApplyPlayerView_PhaseState(t *testing.T) {
	t.Parallel()

	v := gamestate.NewView()
	pv := &snapshot.PlayerView{
		Phase: snapshot.Phase{
			Type: snapshot.PhaseConquer,
			State: snapshot.ConquerPhaseState{
				AttackingRegionID: "a1",
				DefendingRegionID: "a2",
				MinTroopsToMove:   2,
			},
		},
		Mission: snapshot.PlayerMission{
			Type:   snapshot.MissionTwentyFourTerritories,
			Detail: snapshot.TwentyFourTerritoriesMission{},
		},
	}

	err := v.Apply(makePlayerViewMsg(t, pv))
	require.NoError(t, err)

	snap := v.Snapshot()
	require.NotNil(t, snap.PlayerView)
	assert.Equal(t, snapshot.PhaseConquer, snap.PlayerView.Phase.Type)

	state, ok := snap.PlayerView.Phase.State.(snapshot.ConquerPhaseState)
	require.True(t, ok)
	assert.Equal(t, "a1", state.AttackingRegionID)
	assert.Equal(t, "a2", state.DefendingRegionID)
	assert.Equal(t, int64(2), state.MinTroopsToMove)
}

func TestView_ApplyPlayerConnection(t *testing.T) {
	t.Parallel()

	v := gamestate.NewView()

	err := v.Apply(gamestate.WSMessage{Type: "playerConnection", Payload: json.RawMessage(`{}`)})
	require.NoError(t, err)

	// playerConnection should still trigger the updated channel.
	snap := v.Snapshot()
	assert.Nil(t, snap.PlayerView)
}

func TestView_ApplyUnknownType(t *testing.T) {
	t.Parallel()

	v := gamestate.NewView()

	// Unknown types are logged and ignored (no error).
	err := v.Apply(gamestate.WSMessage{Type: "bogus", Payload: json.RawMessage(`{}`)})
	require.NoError(t, err)
}

func TestView_ApplyInvalidJSON(t *testing.T) {
	t.Parallel()

	v := gamestate.NewView()

	err := v.Apply(gamestate.WSMessage{Type: "playerView", Payload: json.RawMessage(`{bad json`)})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal playerView")
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
	err := v.Apply(gamestate.WSMessage{Type: "playerConnection", Payload: json.RawMessage(`{}`)})
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

	err := v.Apply(gamestate.WSMessage{Type: "playerConnection", Payload: json.RawMessage(`{}`)})
	require.NoError(t, err)

	assert.False(t, v.LastUpdateTime().IsZero())
}

func TestViewSnapshot_MyRegions(t *testing.T) {
	t.Parallel()

	snap := gamestate.ViewSnapshot{
		PlayerView: testPlayerView(),
	}

	assert.Len(t, snap.MyRegions("alice"), 2)
	assert.Len(t, snap.MyRegions("bob"), 1)
	assert.Nil(t, snap.MyRegions("charlie"))
}

func TestViewSnapshot_MyRegions_NilPlayerView(t *testing.T) {
	t.Parallel()

	snap := gamestate.ViewSnapshot{}
	assert.Nil(t, snap.MyRegions("alice"))
}

func TestViewSnapshot_CurrentPhase(t *testing.T) {
	t.Parallel()

	snap := gamestate.ViewSnapshot{
		PlayerView: &snapshot.PlayerView{
			Phase: snapshot.Phase{Type: snapshot.PhaseAttack, State: snapshot.EmptyPhaseState{}},
		},
	}
	assert.Equal(t, snapshot.PhaseAttack, snap.CurrentPhase())
}

func TestViewSnapshot_CurrentPhase_NilPlayerView(t *testing.T) {
	t.Parallel()

	snap := gamestate.ViewSnapshot{}
	assert.Equal(t, snapshot.PhaseType(""), snap.CurrentPhase())
}

func TestViewSnapshot_IsMyTurn(t *testing.T) {
	t.Parallel()

	snap := gamestate.ViewSnapshot{
		PlayerView: &snapshot.PlayerView{
			Game: snapshot.GameMeta{Turn: 3},
			Players: []snapshot.PlayerState{
				{UserID: "alice", Index: 0},
				{UserID: "bob", Index: 1},
				{UserID: "charlie", Index: 2},
				{UserID: "diana", Index: 3},
			},
		},
	}

	// Turn 3 % 4 players = index 3 -> diana.
	assert.True(t, snap.IsMyTurn("diana"))
	assert.False(t, snap.IsMyTurn("alice"))
	assert.False(t, snap.IsMyTurn("bob"))
}

func TestViewSnapshot_IsMyTurn_EdgeCases(t *testing.T) {
	t.Parallel()

	// Nil player view.
	snap := gamestate.ViewSnapshot{}
	assert.False(t, snap.IsMyTurn("anyone"))

	// No players.
	snap = gamestate.ViewSnapshot{
		PlayerView: &snapshot.PlayerView{
			Game: snapshot.GameMeta{Turn: 1},
		},
	}
	assert.False(t, snap.IsMyTurn("anyone"))
}

func TestViewSnapshot_IsGameOver(t *testing.T) {
	t.Parallel()

	snap := gamestate.ViewSnapshot{
		PlayerView: &snapshot.PlayerView{
			Game: snapshot.GameMeta{WinnerUserID: "alice"},
		},
	}
	assert.True(t, snap.IsGameOver())

	snap = gamestate.ViewSnapshot{
		PlayerView: &snapshot.PlayerView{},
	}
	assert.False(t, snap.IsGameOver())

	snap = gamestate.ViewSnapshot{}
	assert.False(t, snap.IsGameOver())
}

func TestViewSnapshot_Cards(t *testing.T) {
	t.Parallel()

	snap := gamestate.ViewSnapshot{
		PlayerView: testPlayerView(),
	}

	cards := snap.Cards()
	assert.Len(t, cards, 2)
	assert.Equal(t, snapshot.CardInfantry, cards[0].Type)
	assert.Equal(t, snapshot.CardCavalry, cards[1].Type)
}

func TestViewSnapshot_Cards_NilPlayerView(t *testing.T) {
	t.Parallel()

	snap := gamestate.ViewSnapshot{}
	assert.Nil(t, snap.Cards())
}

func TestViewSnapshot_MyMission(t *testing.T) {
	t.Parallel()

	snap := gamestate.ViewSnapshot{
		PlayerView: testPlayerView(),
	}

	mission := snap.MyMission()
	assert.Equal(t, snapshot.MissionTwentyFourTerritories, mission.Type)
}

func TestViewSnapshot_MyMission_NilPlayerView(t *testing.T) {
	t.Parallel()

	snap := gamestate.ViewSnapshot{}
	mission := snap.MyMission()
	assert.Equal(t, snapshot.MissionType(""), mission.Type)
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

			pv := &snapshot.PlayerView{
				Game: snapshot.GameMeta{Turn: int64(idx)},
			}
			data, marshalErr := json.Marshal(pv)
			assert.NoError(t, marshalErr)
			_ = v.Apply(gamestate.WSMessage{Type: "playerView", Payload: data})
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
