package snapshot_test

import (
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/snapshot"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/game/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

const testGameID = int64(42)

func gameContext(t *testing.T) ctx.GameContext {
	t.Helper()

	return ctx.WithGameID(
		ctx.WithUserID(
			ctx.WithSpan(t.Context(), noop.Span{}),
			"test-user",
		),
		testGameID,
	)
}

func TestGetPublicSnapshot_DeployPhase(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	svc := snapshot.NewService(querier)

	gameCtx := gameContext(t)

	gameRow := sqlc.GetGameRow{
		ID:           testGameID,
		CurrentPhase: sqlc.GamePhaseTypeDEPLOY,
		Turn:         1,
		WinnerUserID: pgtype.Text{},
	}
	deployableTroops := int64(7)
	regions := []sqlc.GetRegionsByGameRow{
		{ID: 1, ExternalReference: "alaska", Troops: 3, UserID: "user-1"},
		{ID: 2, ExternalReference: "brazil", Troops: 5, UserID: "user-2"},
	}
	players := []sqlc.GetPlayersStateRow{
		{UserID: "user-1", Name: "Alice", TurnIndex: 0, CardCount: 2, RegionCount: 1},
		{UserID: "user-2", Name: "Bob", TurnIndex: 1, CardCount: 1, RegionCount: 1},
	}

	querier.EXPECT().GetGame(gameCtx, testGameID).Return(gameRow, nil)
	querier.EXPECT().GetDeployableTroops(gameCtx, testGameID).Return(deployableTroops, nil)
	querier.EXPECT().GetRegionsByGame(gameCtx, testGameID).Return(regions, nil)
	querier.EXPECT().GetPlayersState(gameCtx, testGameID).Return(players, nil)

	result, err := svc.GetPublicSnapshot(gameCtx)

	require.NoError(t, err)
	require.Equal(t, gameRow, result.Game)
	require.Equal(t, sqlc.GamePhaseTypeDEPLOY, result.Phase.Type)
	require.NotNil(t, result.Phase.DeployState)
	require.Equal(t, deployableTroops, result.Phase.DeployState.DeployableTroops)
	require.Nil(t, result.Phase.ConquerState)
	require.Equal(t, regions, result.Board)
	require.Equal(t, players, result.Players)
}

func TestGetPublicSnapshot_ConquerPhase(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	svc := snapshot.NewService(querier)

	gameCtx := gameContext(t)

	gameRow := sqlc.GetGameRow{
		ID:           testGameID,
		CurrentPhase: sqlc.GamePhaseTypeCONQUER,
		Turn:         3,
		WinnerUserID: pgtype.Text{},
	}
	conquerState := sqlc.GetConquerPhaseStateRow{
		SourceRegion:  "alaska",
		TargetRegion:  "kamchatka",
		MinimumTroops: 2,
	}
	regions := []sqlc.GetRegionsByGameRow{
		{ID: 1, ExternalReference: "alaska", Troops: 5, UserID: "user-1"},
	}
	players := []sqlc.GetPlayersStateRow{
		{UserID: "user-1", Name: "Alice", TurnIndex: 0, CardCount: 0, RegionCount: 1},
	}

	querier.EXPECT().GetGame(gameCtx, testGameID).Return(gameRow, nil)
	querier.EXPECT().GetConquerPhaseState(gameCtx, testGameID).Return(conquerState, nil)
	querier.EXPECT().GetRegionsByGame(gameCtx, testGameID).Return(regions, nil)
	querier.EXPECT().GetPlayersState(gameCtx, testGameID).Return(players, nil)

	result, err := svc.GetPublicSnapshot(gameCtx)

	require.NoError(t, err)
	require.Equal(t, gameRow, result.Game)
	require.Equal(t, sqlc.GamePhaseTypeCONQUER, result.Phase.Type)
	require.Nil(t, result.Phase.DeployState)
	require.NotNil(t, result.Phase.ConquerState)
	require.Equal(t, conquerState, *result.Phase.ConquerState)
	require.Equal(t, regions, result.Board)
	require.Equal(t, players, result.Players)
}

func TestGetPublicSnapshot_AttackPhase(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	svc := snapshot.NewService(querier)

	gameCtx := gameContext(t)

	gameRow := sqlc.GetGameRow{
		ID:           testGameID,
		CurrentPhase: sqlc.GamePhaseTypeATTACK,
		Turn:         2,
		WinnerUserID: pgtype.Text{},
	}
	regions := []sqlc.GetRegionsByGameRow{
		{ID: 1, ExternalReference: "alaska", Troops: 3, UserID: "user-1"},
	}
	players := []sqlc.GetPlayersStateRow{
		{UserID: "user-1", Name: "Alice", TurnIndex: 0, CardCount: 0, RegionCount: 1},
	}

	querier.EXPECT().GetGame(gameCtx, testGameID).Return(gameRow, nil)
	querier.EXPECT().GetRegionsByGame(gameCtx, testGameID).Return(regions, nil)
	querier.EXPECT().GetPlayersState(gameCtx, testGameID).Return(players, nil)

	result, err := svc.GetPublicSnapshot(gameCtx)

	require.NoError(t, err)
	require.Equal(t, gameRow, result.Game)
	require.Equal(t, sqlc.GamePhaseTypeATTACK, result.Phase.Type)
	require.Nil(t, result.Phase.DeployState)
	require.Nil(t, result.Phase.ConquerState)
	require.Equal(t, regions, result.Board)
	require.Equal(t, players, result.Players)
}

func TestGetPublicSnapshot_CardsPhase(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	svc := snapshot.NewService(querier)

	gameCtx := gameContext(t)

	gameRow := sqlc.GetGameRow{
		ID:           testGameID,
		CurrentPhase: sqlc.GamePhaseTypeCARDS,
		Turn:         1,
		WinnerUserID: pgtype.Text{},
	}
	regions := []sqlc.GetRegionsByGameRow{}
	players := []sqlc.GetPlayersStateRow{}

	querier.EXPECT().GetGame(gameCtx, testGameID).Return(gameRow, nil)
	querier.EXPECT().GetRegionsByGame(gameCtx, testGameID).Return(regions, nil)
	querier.EXPECT().GetPlayersState(gameCtx, testGameID).Return(players, nil)

	result, err := svc.GetPublicSnapshot(gameCtx)

	require.NoError(t, err)
	require.Equal(t, sqlc.GamePhaseTypeCARDS, result.Phase.Type)
	require.Nil(t, result.Phase.DeployState)
	require.Nil(t, result.Phase.ConquerState)
}

func TestGetPublicSnapshot_ReinforcePhase(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	svc := snapshot.NewService(querier)

	gameCtx := gameContext(t)

	gameRow := sqlc.GetGameRow{
		ID:           testGameID,
		CurrentPhase: sqlc.GamePhaseTypeREINFORCE,
		Turn:         4,
		WinnerUserID: pgtype.Text{},
	}
	regions := []sqlc.GetRegionsByGameRow{}
	players := []sqlc.GetPlayersStateRow{}

	querier.EXPECT().GetGame(gameCtx, testGameID).Return(gameRow, nil)
	querier.EXPECT().GetRegionsByGame(gameCtx, testGameID).Return(regions, nil)
	querier.EXPECT().GetPlayersState(gameCtx, testGameID).Return(players, nil)

	result, err := svc.GetPublicSnapshot(gameCtx)

	require.NoError(t, err)
	require.Equal(t, sqlc.GamePhaseTypeREINFORCE, result.Phase.Type)
	require.Nil(t, result.Phase.DeployState)
	require.Nil(t, result.Phase.ConquerState)
}

func TestGetPublicSnapshot_GameError(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	svc := snapshot.NewService(querier)

	gameCtx := gameContext(t)

	gameErr := errors.New("db connection failed")
	querier.EXPECT().GetGame(gameCtx, testGameID).Return(sqlc.GetGameRow{}, gameErr)

	result, err := svc.GetPublicSnapshot(gameCtx)

	require.Nil(t, result)
	require.ErrorIs(t, err, gameErr)
	require.ErrorContains(t, err, "getting game")
}

func TestGetPublicSnapshot_DeployTroopsError(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	svc := snapshot.NewService(querier)

	gameCtx := gameContext(t)

	gameRow := sqlc.GetGameRow{
		ID:           testGameID,
		CurrentPhase: sqlc.GamePhaseTypeDEPLOY,
		Turn:         1,
		WinnerUserID: pgtype.Text{},
	}
	deployErr := errors.New("deploy query failed")

	querier.EXPECT().GetGame(gameCtx, testGameID).Return(gameRow, nil)
	querier.EXPECT().GetDeployableTroops(gameCtx, testGameID).Return(int64(0), deployErr)

	result, err := svc.GetPublicSnapshot(gameCtx)

	require.Nil(t, result)
	require.ErrorIs(t, err, deployErr)
	require.ErrorContains(t, err, "getting deploy phase state")
}

func TestGetPublicSnapshot_ConquerStateError(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	svc := snapshot.NewService(querier)

	gameCtx := gameContext(t)

	gameRow := sqlc.GetGameRow{
		ID:           testGameID,
		CurrentPhase: sqlc.GamePhaseTypeCONQUER,
		Turn:         3,
		WinnerUserID: pgtype.Text{},
	}
	conquerErr := errors.New("conquer query failed")

	querier.EXPECT().GetGame(gameCtx, testGameID).Return(gameRow, nil)
	querier.EXPECT().
		GetConquerPhaseState(gameCtx, testGameID).
		Return(sqlc.GetConquerPhaseStateRow{}, conquerErr)

	result, err := svc.GetPublicSnapshot(gameCtx)

	require.Nil(t, result)
	require.ErrorIs(t, err, conquerErr)
	require.ErrorContains(t, err, "getting conquer phase state")
}

func TestGetPublicSnapshot_RegionsError(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	svc := snapshot.NewService(querier)

	gameCtx := gameContext(t)

	gameRow := sqlc.GetGameRow{
		ID:           testGameID,
		CurrentPhase: sqlc.GamePhaseTypeATTACK,
		Turn:         2,
		WinnerUserID: pgtype.Text{},
	}
	regionsErr := errors.New("regions query failed")

	querier.EXPECT().GetGame(gameCtx, testGameID).Return(gameRow, nil)
	querier.EXPECT().GetRegionsByGame(gameCtx, testGameID).Return(nil, regionsErr)

	result, err := svc.GetPublicSnapshot(gameCtx)

	require.Nil(t, result)
	require.ErrorIs(t, err, regionsErr)
	require.ErrorContains(t, err, "getting regions")
}

func TestGetPublicSnapshot_PlayersError(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	svc := snapshot.NewService(querier)

	gameCtx := gameContext(t)

	gameRow := sqlc.GetGameRow{
		ID:           testGameID,
		CurrentPhase: sqlc.GamePhaseTypeATTACK,
		Turn:         2,
		WinnerUserID: pgtype.Text{},
	}
	regions := []sqlc.GetRegionsByGameRow{
		{ID: 1, ExternalReference: "alaska", Troops: 3, UserID: "user-1"},
	}
	playersErr := errors.New("players query failed")

	querier.EXPECT().GetGame(gameCtx, testGameID).Return(gameRow, nil)
	querier.EXPECT().GetRegionsByGame(gameCtx, testGameID).Return(regions, nil)
	querier.EXPECT().GetPlayersState(gameCtx, testGameID).Return(nil, playersErr)

	result, err := svc.GetPublicSnapshot(gameCtx)

	require.Nil(t, result)
	require.ErrorIs(t, err, playersErr)
	require.ErrorContains(t, err, "getting players state")
}

func TestGetPrivateSnapshotsByUser_MultiplePlayersPartitioned(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	svc := snapshot.NewService(querier)

	gameCtx := gameContext(t)

	players := []sqlc.GamePlayer{
		{ID: 10, GameID: testGameID, Name: "Alice", UserID: "user-alice", TurnIndex: 0},
		{ID: 20, GameID: testGameID, Name: "Bob", UserID: "user-bob", TurnIndex: 1},
		{ID: 30, GameID: testGameID, Name: "Carol", UserID: "user-carol", TurnIndex: 2},
		{ID: 40, GameID: testGameID, Name: "Dave", UserID: "user-dave", TurnIndex: 3},
	}
	cards := []sqlc.GetAllCardsForGameRow{
		{
			ID:       1,
			CardType: sqlc.GameCardTypeINFANTRY,
			Region:   pgtype.Text{String: "alaska", Valid: true},
			PlayerID: pgtype.Int8{Int64: 10, Valid: true},
		},
		{
			ID:       2,
			CardType: sqlc.GameCardTypeCAVALRY,
			Region:   pgtype.Text{String: "brazil", Valid: true},
			PlayerID: pgtype.Int8{Int64: 10, Valid: true},
		},
		{
			ID:       3,
			CardType: sqlc.GameCardTypeARTILLERY,
			Region:   pgtype.Text{String: "congo", Valid: true},
			PlayerID: pgtype.Int8{Int64: 20, Valid: true},
		},
		{
			ID:       4,
			CardType: sqlc.GameCardTypeJOLLY,
			Region:   pgtype.Text{},
			PlayerID: pgtype.Int8{Int64: 30, Valid: true},
		},
	}
	missions := []sqlc.GameMission{
		{ID: 100, PlayerID: 10, Type: sqlc.GameMissionTypeELIMINATEPLAYER},
		{ID: 200, PlayerID: 20, Type: sqlc.GameMissionTypeTWOCONTINENTS},
		{ID: 300, PlayerID: 30, Type: sqlc.GameMissionTypeTWENTYFOURTERRITORIES},
		{ID: 400, PlayerID: 40, Type: sqlc.GameMissionTypeEIGHTEENTERRITORIESTWOTROOPS},
	}

	querier.EXPECT().GetPlayersByGame(gameCtx, testGameID).Return(players, nil)
	querier.EXPECT().GetAllCardsForGame(gameCtx, testGameID).Return(cards, nil)
	querier.EXPECT().GetAllMissionsForGame(gameCtx, testGameID).Return(missions, nil)

	result, err := svc.GetPrivateSnapshotsByUser(gameCtx)

	require.NoError(t, err)
	require.Len(t, result, 4)

	// Alice (player 10): 2 cards
	require.Equal(t, sqlc.GameMissionTypeELIMINATEPLAYER, result["user-alice"].MissionType)
	require.Equal(t, int64(100), result["user-alice"].MissionID)
	require.Len(t, result["user-alice"].Cards, 2)
	require.Equal(
		t,
		sqlc.GetCardsForPlayerRow{
			ID:       1,
			CardType: sqlc.GameCardTypeINFANTRY,
			Region:   pgtype.Text{String: "alaska", Valid: true},
		},
		result["user-alice"].Cards[0],
	)
	require.Equal(
		t,
		sqlc.GetCardsForPlayerRow{
			ID:       2,
			CardType: sqlc.GameCardTypeCAVALRY,
			Region:   pgtype.Text{String: "brazil", Valid: true},
		},
		result["user-alice"].Cards[1],
	)

	// Bob (player 20): 1 card
	require.Equal(t, sqlc.GameMissionTypeTWOCONTINENTS, result["user-bob"].MissionType)
	require.Equal(t, int64(200), result["user-bob"].MissionID)
	require.Len(t, result["user-bob"].Cards, 1)
	require.Equal(
		t,
		sqlc.GetCardsForPlayerRow{
			ID:       3,
			CardType: sqlc.GameCardTypeARTILLERY,
			Region:   pgtype.Text{String: "congo", Valid: true},
		},
		result["user-bob"].Cards[0],
	)

	// Carol (player 30): 1 card (jolly)
	require.Equal(t, sqlc.GameMissionTypeTWENTYFOURTERRITORIES, result["user-carol"].MissionType)
	require.Equal(t, int64(300), result["user-carol"].MissionID)
	require.Len(t, result["user-carol"].Cards, 1)
	require.Equal(
		t,
		sqlc.GetCardsForPlayerRow{ID: 4, CardType: sqlc.GameCardTypeJOLLY, Region: pgtype.Text{}},
		result["user-carol"].Cards[0],
	)

	// Dave (player 40): mission but no cards
	require.Equal(
		t,
		sqlc.GameMissionTypeEIGHTEENTERRITORIESTWOTROOPS,
		result["user-dave"].MissionType,
	)
	require.Equal(t, int64(400), result["user-dave"].MissionID)
	require.Empty(t, result["user-dave"].Cards)
}

func TestGetPrivateSnapshotsByUser_PlayerWithNoCards(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	svc := snapshot.NewService(querier)

	gameCtx := gameContext(t)

	players := []sqlc.GamePlayer{
		{ID: 10, GameID: testGameID, Name: "Alice", UserID: "user-alice", TurnIndex: 0},
	}
	cards := []sqlc.GetAllCardsForGameRow{} // no owned cards in the game
	missions := []sqlc.GameMission{
		{ID: 100, PlayerID: 10, Type: sqlc.GameMissionTypeTWOCONTINENTSPLUSONE},
	}

	querier.EXPECT().GetPlayersByGame(gameCtx, testGameID).Return(players, nil)
	querier.EXPECT().GetAllCardsForGame(gameCtx, testGameID).Return(cards, nil)
	querier.EXPECT().GetAllMissionsForGame(gameCtx, testGameID).Return(missions, nil)

	result, err := svc.GetPrivateSnapshotsByUser(gameCtx)

	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, sqlc.GameMissionTypeTWOCONTINENTSPLUSONE, result["user-alice"].MissionType)
	require.Equal(t, int64(100), result["user-alice"].MissionID)
	require.Empty(t, result["user-alice"].Cards)
}

func TestGetPrivateSnapshotsByUser_PlayersError(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	svc := snapshot.NewService(querier)

	gameCtx := gameContext(t)

	playersErr := errors.New("players query failed")
	querier.EXPECT().GetPlayersByGame(gameCtx, testGameID).Return(nil, playersErr)

	result, err := svc.GetPrivateSnapshotsByUser(gameCtx)

	require.Nil(t, result)
	require.ErrorIs(t, err, playersErr)
	require.ErrorContains(t, err, "getting players by game")
}

func TestGetPrivateSnapshotsByUser_CardsError(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	svc := snapshot.NewService(querier)

	gameCtx := gameContext(t)

	players := []sqlc.GamePlayer{
		{ID: 10, GameID: testGameID, Name: "Alice", UserID: "user-alice", TurnIndex: 0},
	}
	cardsErr := errors.New("cards query failed")

	querier.EXPECT().GetPlayersByGame(gameCtx, testGameID).Return(players, nil)
	querier.EXPECT().GetAllCardsForGame(gameCtx, testGameID).Return(nil, cardsErr)

	result, err := svc.GetPrivateSnapshotsByUser(gameCtx)

	require.Nil(t, result)
	require.ErrorIs(t, err, cardsErr)
	require.ErrorContains(t, err, "getting all cards for game")
}

func TestGetPrivateSnapshotsByUser_MissionsError(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	svc := snapshot.NewService(querier)

	gameCtx := gameContext(t)

	players := []sqlc.GamePlayer{
		{ID: 10, GameID: testGameID, Name: "Alice", UserID: "user-alice", TurnIndex: 0},
	}
	cards := []sqlc.GetAllCardsForGameRow{}
	missionsErr := errors.New("missions query failed")

	querier.EXPECT().GetPlayersByGame(gameCtx, testGameID).Return(players, nil)
	querier.EXPECT().GetAllCardsForGame(gameCtx, testGameID).Return(cards, nil)
	querier.EXPECT().GetAllMissionsForGame(gameCtx, testGameID).Return(nil, missionsErr)

	result, err := svc.GetPrivateSnapshotsByUser(gameCtx)

	require.Nil(t, result)
	require.ErrorIs(t, err, missionsErr)
	require.ErrorContains(t, err, "getting all missions for game")
}

func TestGetPrivateSnapshotsByUser_EmptyGame(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	svc := snapshot.NewService(querier)

	gameCtx := gameContext(t)

	querier.EXPECT().GetPlayersByGame(gameCtx, testGameID).Return([]sqlc.GamePlayer{}, nil)
	querier.EXPECT().
		GetAllCardsForGame(gameCtx, testGameID).
		Return([]sqlc.GetAllCardsForGameRow{}, nil)
	querier.EXPECT().GetAllMissionsForGame(gameCtx, testGameID).Return([]sqlc.GameMission{}, nil)

	result, err := svc.GetPrivateSnapshotsByUser(gameCtx)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Empty(t, result)
}
