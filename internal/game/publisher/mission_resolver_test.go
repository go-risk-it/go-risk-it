package consumers_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/mission"
	consumers "github.com/go-risk-it/go-risk-it/internal/game/publisher"
	mockMission "github.com/go-risk-it/go-risk-it/mocks/internal_/game/logic/mission"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBuildMissionResolver_TwoContinents(t *testing.T) {
	t.Parallel()

	missionSvc := mockMission.NewService(t)
	missionCtrl := consumers.NewMissionController(missionSvc)

	missionSvc.EXPECT().
		GetTwoContinentsMission(mock.Anything, int64(10)).
		Return(mission.TwoContinentsMission{
			Continent1: "Europe",
			Continent2: "Asia",
		}, nil)

	resolver := consumers.BuildMissionResolver(missionCtrl)
	gameCtx := testGameContext()

	result, err := resolver(gameCtx, sqlc.GameMissionTypeTWOCONTINENTS, 10)
	require.NoError(t, err)

	state, ok := result.(messaging.MissionState[messaging.TwoContinentsMission])
	require.True(t, ok)
	require.Equal(t, messaging.TwoContinents, state.Type)
	require.Equal(t, "Europe", state.Details.Continent1)
	require.Equal(t, "Asia", state.Details.Continent2)
}

func TestBuildMissionResolver_TwoContinentsPlusOne(t *testing.T) {
	t.Parallel()

	missionSvc := mockMission.NewService(t)
	missionCtrl := consumers.NewMissionController(missionSvc)

	missionSvc.EXPECT().
		GetTwoContinentsPlusOneMission(mock.Anything, int64(50)).
		Return(mission.TwoContinentsPlusOneMission{
			Continent1: "Africa",
			Continent2: "SouthAmerica",
		}, nil)

	resolver := consumers.BuildMissionResolver(missionCtrl)
	gameCtx := testGameContext()

	result, err := resolver(gameCtx, sqlc.GameMissionTypeTWOCONTINENTSPLUSONE, 50)
	require.NoError(t, err)

	state, ok := result.(messaging.MissionState[messaging.TwoContinentsPlusOneMission])
	require.True(t, ok)
	require.Equal(t, messaging.TwoContinentsPlusOne, state.Type)
	require.Equal(t, "Africa", state.Details.Continent1)
	require.Equal(t, "SouthAmerica", state.Details.Continent2)
}

func TestBuildMissionResolver_EliminatePlayer(t *testing.T) {
	t.Parallel()

	missionSvc := mockMission.NewService(t)
	missionCtrl := consumers.NewMissionController(missionSvc)

	missionSvc.EXPECT().
		GetEliminatePlayerMission(mock.Anything, int64(20)).
		Return("target-user", nil)

	resolver := consumers.BuildMissionResolver(missionCtrl)
	gameCtx := testGameContext()

	result, err := resolver(gameCtx, sqlc.GameMissionTypeELIMINATEPLAYER, 20)
	require.NoError(t, err)

	state, ok := result.(messaging.MissionState[messaging.EliminatePlayerMission])
	require.True(t, ok)
	require.Equal(t, messaging.EliminatePlayer, state.Type)
	require.Equal(t, "target-user", state.Details.TargetUserID)
}

func TestBuildMissionResolver_EighteenTerritories(t *testing.T) {
	t.Parallel()

	missionSvc := mockMission.NewService(t)
	missionCtrl := consumers.NewMissionController(missionSvc)

	// Static mission -- service is not called.
	resolver := consumers.BuildMissionResolver(missionCtrl)
	gameCtx := testGameContext()

	result, err := resolver(gameCtx, sqlc.GameMissionTypeEIGHTEENTERRITORIESTWOTROOPS, 30)
	require.NoError(t, err)

	state, ok := result.(messaging.MissionState[messaging.EighteenTerritoriesTwoTroopsMission])
	require.True(t, ok)
	require.Equal(t, messaging.EighteenTerritoriesTwoTroops, state.Type)
}

func TestBuildMissionResolver_TwentyFourTerritories(t *testing.T) {
	t.Parallel()

	missionSvc := mockMission.NewService(t)
	missionCtrl := consumers.NewMissionController(missionSvc)

	// Static mission -- service is not called.
	resolver := consumers.BuildMissionResolver(missionCtrl)
	gameCtx := testGameContext()

	result, err := resolver(gameCtx, sqlc.GameMissionTypeTWENTYFOURTERRITORIES, 40)
	require.NoError(t, err)

	state, ok := result.(messaging.MissionState[messaging.TwentyFourTerritoriesMission])
	require.True(t, ok)
	require.Equal(t, messaging.TwentyFourTerritories, state.Type)
}

func TestBuildMissionResolver_UnknownType(t *testing.T) {
	t.Parallel()

	missionSvc := mockMission.NewService(t)
	missionCtrl := consumers.NewMissionController(missionSvc)

	resolver := consumers.BuildMissionResolver(missionCtrl)
	gameCtx := testGameContext()

	_, err := resolver(gameCtx, "INVENTED_TYPE", 99)
	require.Error(t, err)
	require.ErrorContains(t, err, "unknown mission type")
}

func TestBuildMissionResolver_NonGameContext(t *testing.T) {
	t.Parallel()

	missionSvc := mockMission.NewService(t)
	missionCtrl := consumers.NewMissionController(missionSvc)

	resolver := consumers.BuildMissionResolver(missionCtrl)

	_, err := resolver(context.Background(), sqlc.GameMissionTypeTWOCONTINENTS, 10)
	require.Error(t, err)
	require.ErrorContains(t, err, "mission resolver requires GameContext")
}
