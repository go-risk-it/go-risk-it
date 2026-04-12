package handlers_test

import (
	"testing"
	"time"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/handlers"
	mockboard "github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/board"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- helpers ---------------------------------------------------------------

func regionStates(ownership map[string]string) []snapshot.RegionState {
	states := make([]snapshot.RegionState, 0, len(ownership))
	for region, owner := range ownership {
		states = append(states, snapshot.RegionState{ID: region, OwnerID: owner, Troops: 5})
	}

	return states
}

func moveCompletedAttack(
	prevRegions []snapshot.RegionState,
	curRegions []snapshot.RegionState,
) *gameevt.MoveCompleted {
	return gameevt.NewMoveCompleted(
		testGameID, testAttacker, time.Now(),
		gameapi.GamePhaseTypeATTACK,
		testTurn,
		gameapi.GamePhaseTypeATTACK,
		gameapi.GamePhaseTypeATTACK,
		false,
		&snapshot.GameSnapshot{
			Game:    snapshot.GameMeta{ID: testGameID, Turn: testTurn},
			Regions: curRegions,
		},
		nil,
		prevRegions,
	)
}

func setupHeadlinesDetector(
	t *testing.T,
	continentDefs map[string][]string,
) *reentrantBus {
	t.Helper()

	bus := newReentrantBus()
	continents := testContinents(t, continentDefs)

	boardSvc := mockboard.NewService(t)
	boardSvc.EXPECT().GetContinents(mock.Anything).Return(continents, nil).Maybe()

	handlers.RegisterHeadlinesDetector(handlers.HeadlinesDetectorParams{
		Sub:   bus,
		Pub:   bus,
		Board: boardSvc,
	})

	return bus
}

// --- tests -----------------------------------------------------------------

func TestHeadlinesDetectorV2_IgnoresNonAttackMoves(t *testing.T) {
	t.Parallel()

	bus := setupHeadlinesDetector(t, defaultContinents())

	event := gameevt.NewMoveCompleted(
		testGameID, testAttacker, fixedTime,
		gameapi.GamePhaseTypeDEPLOY,
		testTurn,
		gameapi.GamePhaseTypeDEPLOY,
		gameapi.GamePhaseTypeDEPLOY,
		false, nil, nil, nil,
	)

	bus.Emit(gameCtx(), event)

	allEvents := bus.allEvents()
	require.Len(t, allEvents, 1)
	require.Equal(t, gameevt.TypeMoveCompleted, allEvents[0].EventType())
}

func TestHeadlinesDetectorV2_IgnoresAttackWithoutConquest(t *testing.T) {
	t.Parallel()

	bus := setupHeadlinesDetector(t, defaultContinents())

	prev := regionStates(map[string]string{
		"france": testAttacker, "germany": testDefender, "italy": testDefender,
		"china": testDefender, "japan": testDefender,
	})
	cur := prev // no change — attack did not conquer

	event := moveCompletedAttack(prev, cur)

	bus.Emit(gameCtx(), event)

	allEvents := bus.allEvents()
	require.Len(t, allEvents, 1)
}

func TestHeadlinesDetectorV2_SkipsNilPreviousRegions(t *testing.T) {
	t.Parallel()

	bus := setupHeadlinesDetector(t, defaultContinents())

	event := gameevt.NewMoveCompleted(
		testGameID, testAttacker, fixedTime,
		gameapi.GamePhaseTypeATTACK,
		testTurn,
		gameapi.GamePhaseTypeATTACK,
		gameapi.GamePhaseTypeATTACK,
		false,
		&snapshot.GameSnapshot{
			Game:    snapshot.GameMeta{ID: testGameID},
			Regions: regionStates(map[string]string{"france": testAttacker}),
		},
		nil,
		nil, // no previous regions
	)

	bus.Emit(gameCtx(), event)

	allEvents := bus.allEvents()
	require.Len(t, allEvents, 1)
}

func TestHeadlinesDetectorV2_PlayerEliminated(t *testing.T) {
	t.Parallel()

	bus := setupHeadlinesDetector(t, defaultContinents())

	prev := regionStates(map[string]string{
		"france": testAttacker, "germany": testAttacker,
		"italy": testDefender, // defender's only region
		"china": testAttacker, "japan": testAttacker,
	})
	cur := regionStates(map[string]string{
		"france": testAttacker, "germany": testAttacker,
		"italy": testAttacker, // conquered
		"china": testAttacker, "japan": testAttacker,
	})

	event := moveCompletedAttack(prev, cur)

	bus.Emit(gameCtx(), event)

	eliminated := eventsOfType[*gameevt.PlayerEliminated](bus)
	require.Len(t, eliminated, 1)
	require.Equal(t, testDefender, eliminated[0].EliminatedUserID())
	require.Equal(t, testAttacker, eliminated[0].EliminatorUserID())
}

func TestHeadlinesDetectorV2_ContinentCaptured(t *testing.T) {
	t.Parallel()

	bus := setupHeadlinesDetector(t, defaultContinents())

	prev := regionStates(map[string]string{
		"france": testAttacker, "germany": testAttacker,
		"italy": testDefender, // last europe region
		"china": testDefender, "japan": testDefender,
	})
	cur := regionStates(map[string]string{
		"france": testAttacker, "germany": testAttacker,
		"italy": testAttacker, // now all of europe
		"china": testDefender, "japan": testDefender,
	})

	event := moveCompletedAttack(prev, cur)

	bus.Emit(gameCtx(), event)

	captured := eventsOfType[*gameevt.ContinentCaptured](bus)
	require.Len(t, captured, 1)
	require.Equal(t, "europe", captured[0].ContinentID)
	require.Equal(t, testAttacker, captured[0].UserID())
}

func TestHeadlinesDetectorV2_ContinentLost(t *testing.T) {
	t.Parallel()

	bus := setupHeadlinesDetector(t, defaultContinents())

	prev := regionStates(map[string]string{
		"france": testAttacker, "germany": testAttacker, "italy": testAttacker,
		"china": testDefender, "japan": testDefender, // defender owns all of asia
	})
	cur := regionStates(map[string]string{
		"france": testAttacker, "germany": testAttacker, "italy": testAttacker,
		"china": testAttacker, // conquered
		"japan": testDefender,
	})

	event := moveCompletedAttack(prev, cur)

	bus.Emit(gameCtx(), event)

	lost := eventsOfType[*gameevt.ContinentLost](bus)
	require.Len(t, lost, 1)
	require.Equal(t, "asia", lost[0].ContinentID)
	require.Equal(t, testDefender, lost[0].UserID())
}

func TestHeadlinesDetectorV2_ContinentCapturedAndLost(t *testing.T) {
	t.Parallel()

	continentDefs := map[string][]string{
		"island": {"atoll"},
		"big":    {"north", "south"},
	}

	bus := setupHeadlinesDetector(t, continentDefs)

	prev := regionStates(map[string]string{
		"atoll": testDefender,
		"north": testAttacker, "south": testAttacker,
	})
	cur := regionStates(map[string]string{
		"atoll": testAttacker, // conquered
		"north": testAttacker, "south": testAttacker,
	})

	event := moveCompletedAttack(prev, cur)

	bus.Emit(gameCtx(), event)

	captured := eventsOfType[*gameevt.ContinentCaptured](bus)
	require.Len(t, captured, 1)
	require.Equal(t, "island", captured[0].ContinentID)

	lost := eventsOfType[*gameevt.ContinentLost](bus)
	require.Len(t, lost, 1)
	require.Equal(t, "island", lost[0].ContinentID)
}

func TestHeadlinesDetectorV2_NoEliminationWhenDefenderHasRegions(t *testing.T) {
	t.Parallel()

	bus := setupHeadlinesDetector(t, defaultContinents())

	prev := regionStates(map[string]string{
		"france":  testAttacker,
		"germany": testDefender, "italy": testDefender,
		"china": testDefender, "japan": testDefender,
	})
	cur := regionStates(map[string]string{
		"france":  testAttacker,
		"germany": testAttacker, // conquered, but defender still has italy/china/japan
		"italy":   testDefender,
		"china":   testDefender, "japan": testDefender,
	})

	event := moveCompletedAttack(prev, cur)

	bus.Emit(gameCtx(), event)

	eliminated := eventsOfType[*gameevt.PlayerEliminated](bus)
	require.Empty(t, eliminated)
}
