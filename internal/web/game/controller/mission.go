package controller

import (
	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/mission"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
)

type MissionController struct {
	missionService mission.Service
}

func NewMissionController(missionService mission.Service) *MissionController {
	return &MissionController{
		missionService: missionService,
	}
}

// getMissionState is the shared generic helper for all mission state methods.
// It constructs a MissionState[T] by calling fetchDetails to obtain the details,
// then wrapping them with the given mission type.
// For static missions, fetchDetails ignores its arguments and returns a zero-value T.
func getMissionState[T messaging.MissionDetails](
	gameCtx ctx.GameContext,
	missionID int64,
	missionType messaging.MissionType,
	fetchDetails func(ctx.GameContext, int64) (T, error),
) (messaging.MissionState[T], error) {
	details, err := fetchDetails(gameCtx, missionID)
	if err != nil {
		return messaging.MissionState[T]{}, err
	}

	return messaging.MissionState[T]{
		Type:    missionType,
		Details: details,
	}, nil
}

func (m *MissionController) GetTwoContinentsMission(
	gameCtx ctx.GameContext,
	missionID int64,
) (messaging.MissionState[messaging.TwoContinentsMission], error) {
	return getMissionState(gameCtx, missionID, messaging.TwoContinents,
		func(c ctx.GameContext, id int64) (messaging.TwoContinentsMission, error) {
			result, err := m.missionService.GetTwoContinentsMission(c, id)
			if err != nil {
				return messaging.TwoContinentsMission{}, err
			}

			return messaging.TwoContinentsMission{
				Continent1: result.Continent1,
				Continent2: result.Continent2,
			}, nil
		})
}

func (m *MissionController) GetTwoContinentsPlusOneMission(
	gameCtx ctx.GameContext,
	missionID int64,
) (messaging.MissionState[messaging.TwoContinentsPlusOneMission], error) {
	return getMissionState(gameCtx, missionID, messaging.TwoContinentsPlusOne,
		func(c ctx.GameContext, id int64) (messaging.TwoContinentsPlusOneMission, error) {
			result, err := m.missionService.GetTwoContinentsPlusOneMission(c, id)
			if err != nil {
				return messaging.TwoContinentsPlusOneMission{}, err
			}

			return messaging.TwoContinentsPlusOneMission{
				Continent1: result.Continent1,
				Continent2: result.Continent2,
			}, nil
		})
}

func (m *MissionController) GetEliminatePlayerMission(
	gameCtx ctx.GameContext,
	missionID int64,
) (messaging.MissionState[messaging.EliminatePlayerMission], error) {
	return getMissionState(gameCtx, missionID, messaging.EliminatePlayer,
		func(c ctx.GameContext, id int64) (messaging.EliminatePlayerMission, error) {
			targetUser, err := m.missionService.GetEliminatePlayerMission(c, id)
			if err != nil {
				return messaging.EliminatePlayerMission{}, err
			}

			return messaging.EliminatePlayerMission{
				TargetUserID: targetUser,
			}, nil
		})
}

func (m *MissionController) GetEighteenTerritoriesTwoTroopsMission(
	_ ctx.GameContext,
	_ int64,
) (messaging.MissionState[messaging.EighteenTerritoriesTwoTroopsMission], error) {
	return getMissionState[messaging.EighteenTerritoriesTwoTroopsMission](
		nil, 0, messaging.EighteenTerritoriesTwoTroops,
		func(_ ctx.GameContext, _ int64) (messaging.EighteenTerritoriesTwoTroopsMission, error) {
			return messaging.EighteenTerritoriesTwoTroopsMission{}, nil
		})
}

func (m *MissionController) GetTwentyFourTerritoriesMission(
	_ ctx.GameContext,
	_ int64,
) (messaging.MissionState[messaging.TwentyFourTerritoriesMission], error) {
	return getMissionState[messaging.TwentyFourTerritoriesMission](
		nil, 0, messaging.TwentyFourTerritories,
		func(_ ctx.GameContext, _ int64) (messaging.TwentyFourTerritoriesMission, error) {
			return messaging.TwentyFourTerritoriesMission{}, nil
		})
}
