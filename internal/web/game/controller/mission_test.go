package controller_test

import (
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	ctx2 "github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission"
	missionController "github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	missionMock "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/mission"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func missionGameContext(t *testing.T) ctx2.GameContext {
	t.Helper()

	return ctx2.WithGameID(
		ctx2.WithUserID(
			ctx2.WithSpan(t.Context(), noop.Span{}),
			"testuser",
		),
		42,
	)
}

func TestMissionController_GetTwoContinentsMission(t *testing.T) {
	t.Parallel()

	t.Run("returns mapped mission state on success", func(t *testing.T) {
		t.Parallel()

		svc := missionMock.NewService(t)
		ctrl := missionController.NewMissionController(svc)
		gCtx := missionGameContext(t)
		missionID := int64(7)

		svc.EXPECT().
			GetTwoContinentsMission(gCtx, missionID).
			Return(mission.TwoContinentsMission{
				Continent1: "europe",
				Continent2: "asia",
			}, nil)

		got, err := ctrl.GetTwoContinentsMission(gCtx, missionID)

		require.NoError(t, err)
		require.Equal(t, messaging.MissionState[messaging.TwoContinentsMission]{
			Type: messaging.TwoContinents,
			Details: messaging.TwoContinentsMission{
				Continent1: "europe",
				Continent2: "asia",
			},
		}, got)
	})

	t.Run("propagates service error", func(t *testing.T) {
		t.Parallel()

		svc := missionMock.NewService(t)
		ctrl := missionController.NewMissionController(svc)
		gCtx := missionGameContext(t)
		missionID := int64(7)
		serviceErr := errors.New("db connection failed")

		svc.EXPECT().
			GetTwoContinentsMission(gCtx, missionID).
			Return(mission.TwoContinentsMission{}, serviceErr)

		got, err := ctrl.GetTwoContinentsMission(gCtx, missionID)

		require.ErrorIs(t, err, serviceErr)
		require.Equal(t, messaging.MissionState[messaging.TwoContinentsMission]{}, got)
	})
}

func TestMissionController_GetEighteenTerritoriesTwoTroopsMission(t *testing.T) {
	t.Parallel()

	t.Run("returns static mission state without calling service", func(t *testing.T) {
		t.Parallel()

		svc := missionMock.NewService(t)
		ctrl := missionController.NewMissionController(svc)
		gCtx := missionGameContext(t)
		missionID := int64(99)

		// No service expectations — static methods must not call the service.

		got, err := ctrl.GetEighteenTerritoriesTwoTroopsMission(gCtx, missionID)

		require.NoError(t, err)
		require.Equal(t, messaging.MissionState[messaging.EighteenTerritoriesTwoTroopsMission]{
			Type:    messaging.EighteenTerritoriesTwoTroops,
			Details: messaging.EighteenTerritoriesTwoTroopsMission{},
		}, got)
	})
}

func TestMissionController_GetEliminatePlayerMission(t *testing.T) {
	t.Parallel()

	t.Run("returns mapped mission state on success", func(t *testing.T) {
		t.Parallel()

		svc := missionMock.NewService(t)
		ctrl := missionController.NewMissionController(svc)
		gCtx := missionGameContext(t)
		missionID := int64(13)

		svc.EXPECT().
			GetEliminatePlayerMission(gCtx, missionID).
			Return("target-user-42", nil)

		got, err := ctrl.GetEliminatePlayerMission(gCtx, missionID)

		require.NoError(t, err)
		require.Equal(t, messaging.MissionState[messaging.EliminatePlayerMission]{
			Type: messaging.EliminatePlayer,
			Details: messaging.EliminatePlayerMission{
				TargetUserID: "target-user-42",
			},
		}, got)
	})

	t.Run("propagates service error", func(t *testing.T) {
		t.Parallel()

		svc := missionMock.NewService(t)
		ctrl := missionController.NewMissionController(svc)
		gCtx := missionGameContext(t)
		missionID := int64(13)
		serviceErr := errors.New("player not found")

		svc.EXPECT().
			GetEliminatePlayerMission(gCtx, missionID).
			Return("", serviceErr)

		got, err := ctrl.GetEliminatePlayerMission(gCtx, missionID)

		require.ErrorIs(t, err, serviceErr)
		require.Equal(t, messaging.MissionState[messaging.EliminatePlayerMission]{}, got)
	})
}
