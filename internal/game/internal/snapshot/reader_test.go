package snapshot_test

import (
	"errors"
	"testing"

	apisnapshot "github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/snapshot"
	mocksnapshot "github.com/go-risk-it/go-risk-it/internal/game/testmocks/snapshot"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockMissionQuerier implements snapshot.MissionQuerier for testing.
type mockMissionQuerier struct {
	mock.Mock
}

func newMockMissionQuerier(t *testing.T) *mockMissionQuerier {
	t.Helper()

	m := &mockMissionQuerier{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })

	return m
}

func (m *mockMissionQuerier) GetTwoContinentsMission(
	ctx gamectx.GameContext,
	missionID int64,
) (snapshot.TwoContinentsResult, error) {
	args := m.Called(ctx, missionID)

	return args.Get(0).(snapshot.TwoContinentsResult), args.Error(1)
}

func (m *mockMissionQuerier) GetTwoContinentsPlusOneMission(
	ctx gamectx.GameContext,
	missionID int64,
) (snapshot.TwoContinentsPlusOneResult, error) {
	args := m.Called(ctx, missionID)

	return args.Get(0).(snapshot.TwoContinentsPlusOneResult), args.Error(1)
}

func (m *mockMissionQuerier) GetEliminatePlayerMission(
	ctx gamectx.GameContext,
	missionID int64,
) (string, error) {
	args := m.Called(ctx, missionID)

	return args.String(0), args.Error(1)
}

func newReader(
	t *testing.T,
) (*mocksnapshot.Service, *mockMissionQuerier, snapshot.Reader) {
	t.Helper()

	snapshotSvc := mocksnapshot.NewService(t)
	missionQ := newMockMissionQuerier(t)
	reader := snapshot.NewReader(snapshotSvc, missionQ)

	return snapshotSvc, missionQ, reader
}

func TestReader_GetPublicSnapshot_DeployPhase(t *testing.T) {
	t.Parallel()

	snapshotSvc, _, reader := newReader(t)
	ctx := gameContext(t)

	snapshotSvc.EXPECT().GetPublicSnapshot(mock.Anything).Return(&snapshot.PublicSnapshot{
		Game: sqlc.GetGameRow{
			ID:           testGameID,
			CurrentPhase: sqlc.GamePhaseTypeDEPLOY,
			Turn:         1,
			WinnerUserID: pgtype.Text{},
		},
		Phase: snapshot.PhaseState{
			Type:        sqlc.GamePhaseTypeDEPLOY,
			DeployState: &snapshot.DeployState{DeployableTroops: 7},
		},
		Board: []sqlc.GetRegionsByGameRow{
			{ExternalReference: "alaska", UserID: "user-1", Troops: 5},
		},
		Players: []sqlc.GetPlayersStateRow{
			{UserID: "user-1", Name: "Alice", TurnIndex: 0, CardCount: 2, RegionCount: 1},
		},
	}, nil)

	result, err := reader.GetPublicSnapshot(ctx)

	require.NoError(t, err)
	require.Equal(t, apisnapshot.PhaseDeploy, result.Phase.Type)
	require.IsType(t, apisnapshot.DeployPhaseState{}, result.Phase.State)
	require.Equal(
		t,
		int64(7),
		result.Phase.State.(apisnapshot.DeployPhaseState).DeployableTroops,
	)
}

func TestReader_GetPublicSnapshot_ConquerPhase(t *testing.T) {
	t.Parallel()

	snapshotSvc, _, reader := newReader(t)
	ctx := gameContext(t)

	snapshotSvc.EXPECT().GetPublicSnapshot(mock.Anything).Return(&snapshot.PublicSnapshot{
		Game: sqlc.GetGameRow{
			ID:           testGameID,
			CurrentPhase: sqlc.GamePhaseTypeCONQUER,
			Turn:         3,
			WinnerUserID: pgtype.Text{},
		},
		Phase: snapshot.PhaseState{
			Type: sqlc.GamePhaseTypeCONQUER,
			ConquerState: &sqlc.GetConquerPhaseStateRow{
				SourceRegion:  "alaska",
				TargetRegion:  "kamchatka",
				MinimumTroops: 2,
			},
		},
		Board: []sqlc.GetRegionsByGameRow{
			{ExternalReference: "alaska", UserID: "user-1", Troops: 5},
		},
		Players: []sqlc.GetPlayersStateRow{
			{UserID: "user-1", Name: "Alice", TurnIndex: 0, CardCount: 0, RegionCount: 1},
		},
	}, nil)

	result, err := reader.GetPublicSnapshot(ctx)

	require.NoError(t, err)
	require.Equal(t, apisnapshot.PhaseConquer, result.Phase.Type)
	require.IsType(t, apisnapshot.ConquerPhaseState{}, result.Phase.State)

	conquer := result.Phase.State.(apisnapshot.ConquerPhaseState)
	require.Equal(t, "alaska", conquer.AttackingRegionID)
	require.Equal(t, "kamchatka", conquer.DefendingRegionID)
	require.Equal(t, int64(2), conquer.MinTroopsToMove)
}

func TestReader_GetPublicSnapshot_AttackPhase(t *testing.T) {
	t.Parallel()

	snapshotSvc, _, reader := newReader(t)
	ctx := gameContext(t)

	snapshotSvc.EXPECT().GetPublicSnapshot(mock.Anything).Return(&snapshot.PublicSnapshot{
		Game: sqlc.GetGameRow{
			ID:           testGameID,
			CurrentPhase: sqlc.GamePhaseTypeATTACK,
			Turn:         2,
			WinnerUserID: pgtype.Text{},
		},
		Phase: snapshot.PhaseState{
			Type: sqlc.GamePhaseTypeATTACK,
		},
		Board:   []sqlc.GetRegionsByGameRow{},
		Players: []sqlc.GetPlayersStateRow{},
	}, nil)

	result, err := reader.GetPublicSnapshot(ctx)

	require.NoError(t, err)
	require.Equal(t, apisnapshot.PhaseAttack, result.Phase.Type)
	require.Equal(t, apisnapshot.EmptyPhaseState{}, result.Phase.State)
}

func TestReader_GetPublicSnapshot_CardsPhase(t *testing.T) {
	t.Parallel()

	snapshotSvc, _, reader := newReader(t)
	ctx := gameContext(t)

	snapshotSvc.EXPECT().GetPublicSnapshot(mock.Anything).Return(&snapshot.PublicSnapshot{
		Game: sqlc.GetGameRow{
			ID:           testGameID,
			CurrentPhase: sqlc.GamePhaseTypeCARDS,
			Turn:         1,
			WinnerUserID: pgtype.Text{},
		},
		Phase: snapshot.PhaseState{
			Type: sqlc.GamePhaseTypeCARDS,
		},
		Board:   []sqlc.GetRegionsByGameRow{},
		Players: []sqlc.GetPlayersStateRow{},
	}, nil)

	result, err := reader.GetPublicSnapshot(ctx)

	require.NoError(t, err)
	require.Equal(t, apisnapshot.PhaseCards, result.Phase.Type)
	require.Equal(t, apisnapshot.EmptyPhaseState{}, result.Phase.State)
}

func TestReader_GetPublicSnapshot_ReinforcePhase(t *testing.T) {
	t.Parallel()

	snapshotSvc, _, reader := newReader(t)
	ctx := gameContext(t)

	snapshotSvc.EXPECT().GetPublicSnapshot(mock.Anything).Return(&snapshot.PublicSnapshot{
		Game: sqlc.GetGameRow{
			ID:           testGameID,
			CurrentPhase: sqlc.GamePhaseTypeREINFORCE,
			Turn:         4,
			WinnerUserID: pgtype.Text{},
		},
		Phase: snapshot.PhaseState{
			Type: sqlc.GamePhaseTypeREINFORCE,
		},
		Board:   []sqlc.GetRegionsByGameRow{},
		Players: []sqlc.GetPlayersStateRow{},
	}, nil)

	result, err := reader.GetPublicSnapshot(ctx)

	require.NoError(t, err)
	require.Equal(t, apisnapshot.PhaseReinforce, result.Phase.Type)
	require.Equal(t, apisnapshot.EmptyPhaseState{}, result.Phase.State)
}

func TestReader_GetPublicSnapshot_AllFields(t *testing.T) {
	t.Parallel()

	snapshotSvc, _, reader := newReader(t)
	ctx := gameContext(t)

	snapshotSvc.EXPECT().GetPublicSnapshot(mock.Anything).Return(&snapshot.PublicSnapshot{
		Game: sqlc.GetGameRow{
			ID:           testGameID,
			CurrentPhase: sqlc.GamePhaseTypeDEPLOY,
			Turn:         5,
			WinnerUserID: pgtype.Text{String: "user-1", Valid: true},
		},
		Phase: snapshot.PhaseState{
			Type:        sqlc.GamePhaseTypeDEPLOY,
			DeployState: &snapshot.DeployState{DeployableTroops: 3},
		},
		Board: []sqlc.GetRegionsByGameRow{
			{ExternalReference: "alaska", UserID: "user-1", Troops: 5},
			{ExternalReference: "brazil", UserID: "user-2", Troops: 3},
			{ExternalReference: "congo", UserID: "user-1", Troops: 1},
		},
		Players: []sqlc.GetPlayersStateRow{
			{UserID: "user-1", Name: "Alice", TurnIndex: 0, CardCount: 2, RegionCount: 2},
			{UserID: "user-2", Name: "Bob", TurnIndex: 1, CardCount: 0, RegionCount: 0},
		},
	}, nil)

	result, err := reader.GetPublicSnapshot(ctx)

	require.NoError(t, err)

	// Game meta
	require.Equal(t, testGameID, result.Game.ID)
	require.Equal(t, int64(5), result.Game.Turn)
	require.Equal(t, "user-1", result.Game.WinnerUserID)

	// Regions
	require.Len(t, result.Regions, 3)
	require.Equal(t, "alaska", result.Regions[0].ID)
	require.Equal(t, "user-1", result.Regions[0].OwnerID)
	require.Equal(t, int64(5), result.Regions[0].Troops)
	require.Equal(t, "brazil", result.Regions[1].ID)
	require.Equal(t, "congo", result.Regions[2].ID)

	// Players
	require.Len(t, result.Players, 2)
	require.Equal(t, "user-1", result.Players[0].UserID)
	require.Equal(t, "Alice", result.Players[0].Name)
	require.Equal(t, int64(0), result.Players[0].Index)
	require.Equal(t, int64(2), result.Players[0].CardCount)
	require.Equal(t, apisnapshot.PlayerAlive, result.Players[0].Status)

	// Dead player (region count = 0)
	require.Equal(t, "user-2", result.Players[1].UserID)
	require.Equal(t, apisnapshot.PlayerDead, result.Players[1].Status)
}

func TestReader_GetPublicSnapshot_Error(t *testing.T) {
	t.Parallel()

	snapshotSvc, _, reader := newReader(t)
	ctx := gameContext(t)

	delegateErr := errors.New("snapshot service failed")
	snapshotSvc.EXPECT().GetPublicSnapshot(mock.Anything).Return(nil, delegateErr)

	result, err := reader.GetPublicSnapshot(ctx)

	require.Nil(t, result)
	require.ErrorIs(t, err, delegateErr)
	require.ErrorContains(t, err, "getting public snapshot")
}

func TestReader_GetAllPrivateSnapshots_TwoContinents(t *testing.T) {
	t.Parallel()

	snapshotSvc, missionQ, reader := newReader(t)
	ctx := gameContext(t)

	snapshotSvc.EXPECT().GetPrivateSnapshotsByUser(mock.Anything).Return(
		map[string]*snapshot.PrivateSnapshot{
			"user-alice": {
				Cards: []sqlc.GetCardsForPlayerRow{
					{
						ID:       1,
						CardType: sqlc.GameCardTypeINFANTRY,
						Region:   pgtype.Text{String: "alaska", Valid: true},
					},
					{
						ID:       2,
						CardType: sqlc.GameCardTypeJOLLY,
						Region:   pgtype.Text{},
					},
				},
				MissionType: sqlc.GameMissionTypeTWOCONTINENTS,
				MissionID:   100,
			},
		}, nil)

	missionQ.On("GetTwoContinentsMission", mock.Anything, int64(100)).
		Return(snapshot.TwoContinentsResult{
			Continent1: "europe",
			Continent2: "asia",
		}, nil)

	result, err := reader.GetAllPrivateSnapshots(ctx)

	require.NoError(t, err)
	require.Len(t, result, 1)

	alice := result["user-alice"]
	require.NotNil(t, alice)

	// Cards
	require.Len(t, alice.Cards, 2)
	require.Equal(t, apisnapshot.CardInfantry, alice.Cards[0].Type)
	require.Equal(t, "alaska", alice.Cards[0].Region)
	require.Equal(t, int64(1), alice.Cards[0].ID)
	require.Equal(t, apisnapshot.CardJolly, alice.Cards[1].Type)
	require.Empty(t, alice.Cards[1].Region)

	// Mission
	require.Equal(t, apisnapshot.MissionTwoContinents, alice.Mission.Type)
	require.IsType(t, apisnapshot.TwoContinentsMission{}, alice.Mission.Detail)

	detail := alice.Mission.Detail.(apisnapshot.TwoContinentsMission)
	require.Equal(t, "europe", detail.Continent1)
	require.Equal(t, "asia", detail.Continent2)
}

func TestReader_GetAllPrivateSnapshots_AllMissionTypes(t *testing.T) {
	t.Parallel()

	snapshotSvc, missionQ, reader := newReader(t)
	ctx := gameContext(t)

	snapshotSvc.EXPECT().GetPrivateSnapshotsByUser(mock.Anything).Return(
		map[string]*snapshot.PrivateSnapshot{
			"user-a": {
				MissionType: sqlc.GameMissionTypeTWOCONTINENTSPLUSONE,
				MissionID:   200,
			},
			"user-b": {
				MissionType: sqlc.GameMissionTypeEIGHTEENTERRITORIESTWOTROOPS,
				MissionID:   300,
			},
			"user-c": {
				MissionType: sqlc.GameMissionTypeTWENTYFOURTERRITORIES,
				MissionID:   400,
			},
			"user-d": {
				MissionType: sqlc.GameMissionTypeELIMINATEPLAYER,
				MissionID:   500,
			},
		}, nil)

	missionQ.On("GetTwoContinentsPlusOneMission", mock.Anything, int64(200)).
		Return(snapshot.TwoContinentsPlusOneResult{
			Continent1: "europe",
			Continent2: "south_america",
		}, nil)
	missionQ.On("GetEliminatePlayerMission", mock.Anything, int64(500)).
		Return("user-target", nil)

	result, err := reader.GetAllPrivateSnapshots(ctx)

	require.NoError(t, err)
	require.Len(t, result, 4)

	// TwoContinentsPlusOne
	require.Equal(t, apisnapshot.MissionTwoContinentsPlusOne, result["user-a"].Mission.Type)

	detail := result["user-a"].Mission.Detail.(apisnapshot.TwoContinentsPlusOneMission)
	require.Equal(t, "europe", detail.Continent1)
	require.Equal(t, "south_america", detail.Continent2)

	// EighteenTerritoriesTwoTroops (static -- no mission querier call)
	require.Equal(
		t,
		apisnapshot.MissionEighteenTerritoriesTwoTroops,
		result["user-b"].Mission.Type,
	)
	require.Equal(
		t,
		apisnapshot.EighteenTerritoriesTwoTroopsMission{},
		result["user-b"].Mission.Detail,
	)

	// TwentyFourTerritories (static -- no mission querier call)
	require.Equal(
		t,
		apisnapshot.MissionTwentyFourTerritories,
		result["user-c"].Mission.Type,
	)
	require.Equal(
		t,
		apisnapshot.TwentyFourTerritoriesMission{},
		result["user-c"].Mission.Detail,
	)

	// EliminatePlayer
	require.Equal(t, apisnapshot.MissionEliminatePlayer, result["user-d"].Mission.Type)

	elimDetail := result["user-d"].Mission.Detail.(apisnapshot.EliminatePlayerMission)
	require.Equal(t, "user-target", elimDetail.TargetUserID)
}

func TestReader_GetAllPrivateSnapshots_AllCardTypes(t *testing.T) {
	t.Parallel()

	snapshotSvc, _, reader := newReader(t)
	ctx := gameContext(t)

	snapshotSvc.EXPECT().GetPrivateSnapshotsByUser(mock.Anything).Return(
		map[string]*snapshot.PrivateSnapshot{
			"user-a": {
				Cards: []sqlc.GetCardsForPlayerRow{
					{
						ID:       1,
						CardType: sqlc.GameCardTypeINFANTRY,
						Region:   pgtype.Text{String: "alaska", Valid: true},
					},
					{
						ID:       2,
						CardType: sqlc.GameCardTypeCAVALRY,
						Region:   pgtype.Text{String: "brazil", Valid: true},
					},
					{
						ID:       3,
						CardType: sqlc.GameCardTypeARTILLERY,
						Region:   pgtype.Text{String: "congo", Valid: true},
					},
					{
						ID:       4,
						CardType: sqlc.GameCardTypeJOLLY,
						Region:   pgtype.Text{},
					},
				},
				MissionType: sqlc.GameMissionTypeTWENTYFOURTERRITORIES,
				MissionID:   100,
			},
		}, nil)

	result, err := reader.GetAllPrivateSnapshots(ctx)

	require.NoError(t, err)

	cards := result["user-a"].Cards
	require.Len(t, cards, 4)
	require.Equal(t, apisnapshot.CardInfantry, cards[0].Type)
	require.Equal(t, "alaska", cards[0].Region)
	require.Equal(t, apisnapshot.CardCavalry, cards[1].Type)
	require.Equal(t, "brazil", cards[1].Region)
	require.Equal(t, apisnapshot.CardArtillery, cards[2].Type)
	require.Equal(t, "congo", cards[2].Region)
	require.Equal(t, apisnapshot.CardJolly, cards[3].Type)
	require.Empty(t, cards[3].Region)
}

func TestReader_GetAllPrivateSnapshots_Error(t *testing.T) {
	t.Parallel()

	snapshotSvc, _, reader := newReader(t)
	ctx := gameContext(t)

	delegateErr := errors.New("private snapshot failed")
	snapshotSvc.EXPECT().GetPrivateSnapshotsByUser(mock.Anything).Return(nil, delegateErr)

	result, err := reader.GetAllPrivateSnapshots(ctx)

	require.Nil(t, result)
	require.ErrorIs(t, err, delegateErr)
	require.ErrorContains(t, err, "getting private snapshots")
}

func TestReader_GetAllPrivateSnapshots_MissionResolutionError(t *testing.T) {
	t.Parallel()

	snapshotSvc, missionQ, reader := newReader(t)
	ctx := gameContext(t)

	snapshotSvc.EXPECT().GetPrivateSnapshotsByUser(mock.Anything).Return(
		map[string]*snapshot.PrivateSnapshot{
			"user-alice": {
				MissionType: sqlc.GameMissionTypeTWOCONTINENTS,
				MissionID:   100,
			},
		}, nil)

	missionErr := errors.New("mission DB failed")
	missionQ.On("GetTwoContinentsMission", mock.Anything, int64(100)).
		Return(snapshot.TwoContinentsResult{}, missionErr)

	result, err := reader.GetAllPrivateSnapshots(ctx)

	require.Nil(t, result)
	require.ErrorIs(t, err, missionErr)
	require.ErrorContains(t, err, "resolving mission for user")
}
