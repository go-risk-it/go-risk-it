package signals_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/game/signals"
	mockMission "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/mission"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func testGameContext(t *testing.T) ctx.GameContext {
	t.Helper()

	const gameID = int64(42)

	return ctx.WithGameID(
		ctx.WithUserID(ctx.WithSpan(t.Context(), noop.Span{}), "user-1"),
		gameID,
	)
}

func TestBuildMissionResolver_TwoContinents(t *testing.T) {
	t.Parallel()

	missionSvc := mockMission.NewService(t)
	missionCtrl := controller.NewMissionController(missionSvc)

	missionSvc.EXPECT().
		GetTwoContinentsMission(mock.Anything, int64(10)).
		Return(mission.TwoContinentsMission{
			Continent1: "Europe",
			Continent2: "Asia",
		}, nil)

	resolver := signals.BuildMissionResolver(missionCtrl)
	gameCtx := testGameContext(t)

	raw, err := resolver(gameCtx, sqlc.GameMissionTypeTWOCONTINENTS, 10)
	require.NoError(t, err)

	msg := unmarshalEnvelope(t, raw)
	require.Equal(t, "missionState", msg.Type)

	var data map[string]any
	require.NoError(t, json.Unmarshal(msg.Payload, &data))
	require.Equal(t, "twoContinents", data["type"])

	details, ok := data["details"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Europe", details["continent1"])
	require.Equal(t, "Asia", details["continent2"])
}

func TestBuildMissionResolver_TwoContinentsPlusOne(t *testing.T) {
	t.Parallel()

	missionSvc := mockMission.NewService(t)
	missionCtrl := controller.NewMissionController(missionSvc)

	missionSvc.EXPECT().
		GetTwoContinentsPlusOneMission(mock.Anything, int64(50)).
		Return(mission.TwoContinentsPlusOneMission{
			Continent1: "Africa",
			Continent2: "SouthAmerica",
		}, nil)

	resolver := signals.BuildMissionResolver(missionCtrl)
	gameCtx := testGameContext(t)

	raw, err := resolver(gameCtx, sqlc.GameMissionTypeTWOCONTINENTSPLUSONE, 50)
	require.NoError(t, err)

	msg := unmarshalEnvelope(t, raw)
	require.Equal(t, "missionState", msg.Type)

	var data map[string]any
	require.NoError(t, json.Unmarshal(msg.Payload, &data))
	require.Equal(t, "twoContinentsPlusOne", data["type"])

	details, ok := data["details"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Africa", details["continent1"])
	require.Equal(t, "SouthAmerica", details["continent2"])
}

func TestBuildMissionResolver_EliminatePlayer(t *testing.T) {
	t.Parallel()

	missionSvc := mockMission.NewService(t)
	missionCtrl := controller.NewMissionController(missionSvc)

	missionSvc.EXPECT().
		GetEliminatePlayerMission(mock.Anything, int64(20)).
		Return("target-user", nil)

	resolver := signals.BuildMissionResolver(missionCtrl)
	gameCtx := testGameContext(t)

	raw, err := resolver(gameCtx, sqlc.GameMissionTypeELIMINATEPLAYER, 20)
	require.NoError(t, err)

	msg := unmarshalEnvelope(t, raw)
	require.Equal(t, "missionState", msg.Type)

	var data map[string]any
	require.NoError(t, json.Unmarshal(msg.Payload, &data))
	require.Equal(t, "eliminatePlayer", data["type"])

	details, ok := data["details"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "target-user", details["targetUserId"])
}

func TestBuildMissionResolver_EighteenTerritories(t *testing.T) {
	t.Parallel()

	missionSvc := mockMission.NewService(t)
	missionCtrl := controller.NewMissionController(missionSvc)

	// Static mission — service is not called.
	resolver := signals.BuildMissionResolver(missionCtrl)
	gameCtx := testGameContext(t)

	raw, err := resolver(gameCtx, sqlc.GameMissionTypeEIGHTEENTERRITORIESTWOTROOPS, 30)
	require.NoError(t, err)

	msg := unmarshalEnvelope(t, raw)
	require.Equal(t, "missionState", msg.Type)

	var data map[string]any
	require.NoError(t, json.Unmarshal(msg.Payload, &data))
	require.Equal(t, "eighteenTerritoriesTwoTroops", data["type"])
}

func TestBuildMissionResolver_TwentyFourTerritories(t *testing.T) {
	t.Parallel()

	missionSvc := mockMission.NewService(t)
	missionCtrl := controller.NewMissionController(missionSvc)

	// Static mission — service is not called.
	resolver := signals.BuildMissionResolver(missionCtrl)
	gameCtx := testGameContext(t)

	raw, err := resolver(gameCtx, sqlc.GameMissionTypeTWENTYFOURTERRITORIES, 40)
	require.NoError(t, err)

	msg := unmarshalEnvelope(t, raw)
	require.Equal(t, "missionState", msg.Type)

	var data map[string]any
	require.NoError(t, json.Unmarshal(msg.Payload, &data))
	require.Equal(t, "twentyFourTerritories", data["type"])
}

func TestBuildMissionResolver_UnknownType(t *testing.T) {
	t.Parallel()

	missionSvc := mockMission.NewService(t)
	missionCtrl := controller.NewMissionController(missionSvc)

	resolver := signals.BuildMissionResolver(missionCtrl)
	gameCtx := testGameContext(t)

	_, err := resolver(gameCtx, "INVENTED_TYPE", 99)
	require.Error(t, err)
	require.ErrorContains(t, err, "unknown mission type")
}

func TestBuildMissionResolver_NonGameContext(t *testing.T) {
	t.Parallel()

	missionSvc := mockMission.NewService(t)
	missionCtrl := controller.NewMissionController(missionSvc)

	resolver := signals.BuildMissionResolver(missionCtrl)

	_, err := resolver(context.Background(), sqlc.GameMissionTypeTWOCONTINENTS, 10)
	require.Error(t, err)
	require.ErrorContains(t, err, "mission resolver requires GameContext")
}

// unmarshalEnvelope parses a WS message envelope from raw JSON.
func unmarshalEnvelope(t *testing.T, raw json.RawMessage) struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"data"`
} {
	t.Helper()

	var msg struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(raw, &msg))

	return msg
}
