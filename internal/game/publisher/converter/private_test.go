package converter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/publisher/converter"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func stubMissionResolver(
	val any,
	err error,
) converter.MissionResolver {
	return func(
		_ context.Context,
		_ sqlc.GameMissionType,
		_ int64,
	) (any, error) {
		return val, err
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

	fakeMission := "stub-mission"
	result, err := converter.ConvertPrivateSnapshot(
		t.Context(),
		snap,
		stubMissionResolver(fakeMission, nil),
	)
	require.NoError(t, err)

	// Verify card state
	require.Len(t, result.CardState.Cards, 1)
	require.Equal(t, int64(1), result.CardState.Cards[0].ID)
	require.Equal(t, messaging.Infantry, result.CardState.Cards[0].Type)
	require.Equal(t, "alaska", result.CardState.Cards[0].Region)

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

	result, err := converter.ConvertPrivateSnapshot(
		t.Context(),
		snap,
		stubMissionResolver("stub", nil),
	)
	require.NoError(t, err)

	require.Len(t, result.CardState.Cards, 1)
	require.Equal(t, messaging.Cavalry, result.CardState.Cards[0].Type)
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

	result, err := converter.ConvertPrivateSnapshot(
		t.Context(),
		snap,
		stubMissionResolver("stub", nil),
	)
	require.NoError(t, err)

	require.Len(t, result.CardState.Cards, 1)
	require.Equal(t, messaging.Artillery, result.CardState.Cards[0].Type)
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

	result, err := converter.ConvertPrivateSnapshot(
		t.Context(),
		snap,
		stubMissionResolver("stub", nil),
	)
	require.NoError(t, err)

	require.Len(t, result.CardState.Cards, 1)
	require.Equal(t, messaging.Jolly, result.CardState.Cards[0].Type)
	require.Empty(t, result.CardState.Cards[0].Region)
}

func TestConvertPrivateSnapshot_EmptyCardList(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PrivateSnapshot{
		Cards:       []sqlc.GetCardsForPlayerRow{},
		MissionType: sqlc.GameMissionTypeEIGHTEENTERRITORIESTWOTROOPS,
		MissionID:   300,
	}

	result, err := converter.ConvertPrivateSnapshot(
		t.Context(),
		snap,
		stubMissionResolver("stub", nil),
	)
	require.NoError(t, err)

	require.Empty(t, result.CardState.Cards)
}

func TestConvertPrivateSnapshot_NilCardList(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PrivateSnapshot{
		Cards:       nil,
		MissionType: sqlc.GameMissionTypeELIMINATEPLAYER,
		MissionID:   400,
	}

	result, err := converter.ConvertPrivateSnapshot(
		t.Context(),
		snap,
		stubMissionResolver("stub", nil),
	)
	require.NoError(t, err)

	require.Empty(t, result.CardState.Cards)
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
	) (any, error) {
		capturedType = missionType
		capturedID = missionID

		return "stub", nil
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

	result, err := converter.ConvertPrivateSnapshot(
		t.Context(),
		snap,
		stubMissionResolver("stub", nil),
	)
	require.NoError(t, err)

	require.Len(t, result.CardState.Cards, 4)

	// Verify order preserved
	require.Equal(t, messaging.Infantry, result.CardState.Cards[0].Type)
	require.Equal(t, "alaska", result.CardState.Cards[0].Region)

	require.Equal(t, messaging.Jolly, result.CardState.Cards[3].Type)
	require.Empty(t, result.CardState.Cards[3].Region)
}
