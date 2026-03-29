package converter_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/consumers/converter"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/snapshot"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func stubMissionResolver(
	msg json.RawMessage,
	err error,
) converter.MissionResolver {
	return func(
		_ context.Context,
		_ sqlc.GameMissionType,
		_ int64,
	) (json.RawMessage, error) {
		return msg, err
	}
}

func TestConvertPrivateSnapshot_Infantry(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PrivateSnapshot{
		Cards: []sqlc.GetCardsForPlayerRow{
			{
				ID:       1,
				CardType: sqlc.GameCardTypeINFANTRY,
				Region:   pgtype.Text{String: "alaska", Valid: true},
			},
		},
		MissionType: sqlc.GameMissionTypeTWOCONTINENTS,
		MissionID:   100,
	}

	fakeMission := json.RawMessage(`{"type":"missionState","data":{"type":"twoContinents"}}`)
	result, err := converter.ConvertPrivateSnapshot(
		t.Context(),
		snap,
		stubMissionResolver(fakeMission, nil),
	)
	require.NoError(t, err)

	// Verify card state
	cardMsg := unmarshalMessage(t, result.CardState)
	require.Equal(t, "cardState", cardMsg.Type)

	var cardData map[string]any
	require.NoError(t, json.Unmarshal(cardMsg.Payload, &cardData))

	cards, ok := cardData["cards"].([]any)
	require.True(t, ok)
	require.Len(t, cards, 1)

	card0, ok := cards[0].(map[string]any)
	require.True(t, ok)
	require.InDelta(t, float64(1), card0["id"], 0)
	require.Equal(t, "infantry", card0["type"])
	require.Equal(t, "alaska", card0["region"])

	// Verify mission state is passed through from resolver
	require.Equal(t, fakeMission, result.MissionState)
}

func TestConvertPrivateSnapshot_Cavalry(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PrivateSnapshot{
		Cards: []sqlc.GetCardsForPlayerRow{
			{
				ID:       2,
				CardType: sqlc.GameCardTypeCAVALRY,
				Region:   pgtype.Text{String: "brazil", Valid: true},
			},
		},
		MissionType: sqlc.GameMissionTypeTWOCONTINENTS,
		MissionID:   100,
	}

	fakeMission := json.RawMessage(`{}`)
	result, err := converter.ConvertPrivateSnapshot(
		t.Context(),
		snap,
		stubMissionResolver(fakeMission, nil),
	)
	require.NoError(t, err)

	cardMsg := unmarshalMessage(t, result.CardState)
	var cardData map[string]any
	require.NoError(t, json.Unmarshal(cardMsg.Payload, &cardData))

	cards, ok := cardData["cards"].([]any)
	require.True(t, ok)
	card0, ok := cards[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "cavalry", card0["type"])
}

func TestConvertPrivateSnapshot_Artillery(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PrivateSnapshot{
		Cards: []sqlc.GetCardsForPlayerRow{
			{
				ID:       3,
				CardType: sqlc.GameCardTypeARTILLERY,
				Region:   pgtype.Text{String: "congo", Valid: true},
			},
		},
		MissionType: sqlc.GameMissionTypeTWOCONTINENTS,
		MissionID:   100,
	}

	fakeMission := json.RawMessage(`{}`)
	result, err := converter.ConvertPrivateSnapshot(
		t.Context(),
		snap,
		stubMissionResolver(fakeMission, nil),
	)
	require.NoError(t, err)

	cardMsg := unmarshalMessage(t, result.CardState)
	var cardData map[string]any
	require.NoError(t, json.Unmarshal(cardMsg.Payload, &cardData))

	cards, ok := cardData["cards"].([]any)
	require.True(t, ok)
	card0, ok := cards[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "artillery", card0["type"])
}

func TestConvertPrivateSnapshot_JollyCard(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PrivateSnapshot{
		Cards: []sqlc.GetCardsForPlayerRow{
			{
				ID:       4,
				CardType: sqlc.GameCardTypeJOLLY,
				Region:   pgtype.Text{}, // jolly has no region
			},
		},
		MissionType: sqlc.GameMissionTypeTWENTYFOURTERRITORIES,
		MissionID:   200,
	}

	fakeMission := json.RawMessage(`{}`)
	result, err := converter.ConvertPrivateSnapshot(
		t.Context(),
		snap,
		stubMissionResolver(fakeMission, nil),
	)
	require.NoError(t, err)

	cardMsg := unmarshalMessage(t, result.CardState)
	var cardData map[string]any
	require.NoError(t, json.Unmarshal(cardMsg.Payload, &cardData))

	cards, ok := cardData["cards"].([]any)
	require.True(t, ok)
	card0, ok := cards[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "jolly", card0["type"])
	require.Empty(t, card0["region"])
}

func TestConvertPrivateSnapshot_EmptyCardList(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PrivateSnapshot{
		Cards:       []sqlc.GetCardsForPlayerRow{},
		MissionType: sqlc.GameMissionTypeEIGHTEENTERRITORIESTWOTROOPS,
		MissionID:   300,
	}

	fakeMission := json.RawMessage(`{}`)
	result, err := converter.ConvertPrivateSnapshot(
		t.Context(),
		snap,
		stubMissionResolver(fakeMission, nil),
	)
	require.NoError(t, err)

	cardMsg := unmarshalMessage(t, result.CardState)
	var cardData map[string]any
	require.NoError(t, json.Unmarshal(cardMsg.Payload, &cardData))

	cards, ok := cardData["cards"].([]any)
	require.True(t, ok)
	require.Empty(t, cards)
}

func TestConvertPrivateSnapshot_NilCardList(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PrivateSnapshot{
		Cards:       nil,
		MissionType: sqlc.GameMissionTypeELIMINATEPLAYER,
		MissionID:   400,
	}

	fakeMission := json.RawMessage(`{}`)
	result, err := converter.ConvertPrivateSnapshot(
		t.Context(),
		snap,
		stubMissionResolver(fakeMission, nil),
	)
	require.NoError(t, err)

	cardMsg := unmarshalMessage(t, result.CardState)
	var cardData map[string]any
	require.NoError(t, json.Unmarshal(cardMsg.Payload, &cardData))

	cards, ok := cardData["cards"].([]any)
	require.True(t, ok)
	require.Empty(t, cards)
}

func TestConvertPrivateSnapshot_MissionResolverError(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PrivateSnapshot{
		Cards:       []sqlc.GetCardsForPlayerRow{},
		MissionType: sqlc.GameMissionTypeTWOCONTINENTS,
		MissionID:   100,
	}

	resolverErr := errors.New("mission service unavailable")
	_, err := converter.ConvertPrivateSnapshot(
		t.Context(),
		snap,
		stubMissionResolver(nil, resolverErr),
	)
	require.Error(t, err)
	require.ErrorIs(t, err, resolverErr)
	require.ErrorContains(t, err, "resolving mission state")
}

func TestConvertPrivateSnapshot_MissionResolverReceivesCorrectArgs(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PrivateSnapshot{
		Cards:       []sqlc.GetCardsForPlayerRow{},
		MissionType: sqlc.GameMissionTypeELIMINATEPLAYER,
		MissionID:   999,
	}

	var capturedType sqlc.GameMissionType
	var capturedID int64

	resolver := func(
		_ context.Context,
		missionType sqlc.GameMissionType,
		missionID int64,
	) (json.RawMessage, error) {
		capturedType = missionType
		capturedID = missionID

		return json.RawMessage(`{}`), nil
	}

	_, err := converter.ConvertPrivateSnapshot(t.Context(), snap, resolver)
	require.NoError(t, err)
	require.Equal(t, sqlc.GameMissionTypeELIMINATEPLAYER, capturedType)
	require.Equal(t, int64(999), capturedID)
}

func TestConvertPrivateSnapshot_MultipleCards(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PrivateSnapshot{
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
			{ID: 4, CardType: sqlc.GameCardTypeJOLLY, Region: pgtype.Text{}},
		},
		MissionType: sqlc.GameMissionTypeTWOCONTINENTSPLUSONE,
		MissionID:   500,
	}

	fakeMission := json.RawMessage(`{}`)
	result, err := converter.ConvertPrivateSnapshot(
		t.Context(),
		snap,
		stubMissionResolver(fakeMission, nil),
	)
	require.NoError(t, err)

	cardMsg := unmarshalMessage(t, result.CardState)
	var cardData map[string]any
	require.NoError(t, json.Unmarshal(cardMsg.Payload, &cardData))

	cards, ok := cardData["cards"].([]any) //nolint:varnamelen
	require.True(t, ok)
	require.Len(t, cards, 4)

	// Verify order preserved
	card0, ok := cards[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "infantry", card0["type"])
	require.Equal(t, "alaska", card0["region"])

	card3, ok := cards[3].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "jolly", card3["type"])
	require.Empty(t, card3["region"])
}
