package controller

import (
	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission"
)

type MissionController interface {
	GetTwoContinentsMission(
		ctx ctx.GameContext, missionID int64,
	) (messaging.MissionState[messaging.TwoContinentsMission], error)
	GetTwoContinentsPlusOneMission(
		ctx ctx.GameContext, missionID int64,
	) (messaging.MissionState[messaging.TwoContinentsPlusOneMission], error)
	GetEighteenTerritoriesTwoTroopsMission(
		ctx ctx.GameContext, missionID int64,
	) (messaging.MissionState[messaging.EighteenTerritoriesTwoTroopsMission], error)
	GetTwentyFourTerritoriesMission(
		ctx ctx.GameContext, missionID int64,
	) (messaging.MissionState[messaging.TwentyFourTerritoriesMission], error)
	GetEliminatePlayerMission(
		ctx ctx.GameContext, missionID int64,
	) (messaging.MissionState[messaging.EliminatePlayerMission], error)
}

type MissionControllerImpl struct {
	missionService mission.Service
}

var _ MissionController = (*MissionControllerImpl)(nil)

func NewMissionController(missionService mission.Service) *MissionControllerImpl {
	return &MissionControllerImpl{
		missionService: missionService,
	}
}

func (m *MissionControllerImpl) GetTwoContinentsMission(
	ctx ctx.GameContext,
	missionID int64,
) (messaging.MissionState[messaging.TwoContinentsMission], error) {
	missionDetails, err := m.missionService.GetTwoContinentsMission(ctx, missionID)
	if err != nil {
		return messaging.MissionState[messaging.TwoContinentsMission]{}, err
	}

	return messaging.MissionState[messaging.TwoContinentsMission]{
		Type: messaging.TwoContinents,
		Details: messaging.TwoContinentsMission{
			Continent1: missionDetails.Continent1,
			Continent2: missionDetails.Continent2,
		},
	}, nil
}

func (m *MissionControllerImpl) GetTwoContinentsPlusOneMission(
	ctx ctx.GameContext,
	missionID int64,
) (messaging.MissionState[messaging.TwoContinentsPlusOneMission], error) {
	missionDetails, err := m.missionService.GetTwoContinentsPlusOneMission(ctx, missionID)
	if err != nil {
		return messaging.MissionState[messaging.TwoContinentsPlusOneMission]{}, err
	}

	return messaging.MissionState[messaging.TwoContinentsPlusOneMission]{
		Type: messaging.TwoContinentsPlusOne,
		Details: messaging.TwoContinentsPlusOneMission{
			Continent1: missionDetails.Continent1,
			Continent2: missionDetails.Continent2,
		},
	}, nil
}

func (m *MissionControllerImpl) GetEliminatePlayerMission(
	ctx ctx.GameContext,
	missionID int64,
) (messaging.MissionState[messaging.EliminatePlayerMission], error) {
	targetUser, err := m.missionService.GetEliminatePlayerMission(ctx, missionID)
	if err != nil {
		return messaging.MissionState[messaging.EliminatePlayerMission]{}, err
	}

	return messaging.MissionState[messaging.EliminatePlayerMission]{
		Type: messaging.EliminatePlayer,
		Details: messaging.EliminatePlayerMission{
			TargetUserID: targetUser,
		},
	}, nil
}

func (m *MissionControllerImpl) GetEighteenTerritoriesTwoTroopsMission(
	_ ctx.GameContext,
	_ int64,
) (messaging.MissionState[messaging.EighteenTerritoriesTwoTroopsMission], error) {
	return messaging.MissionState[messaging.EighteenTerritoriesTwoTroopsMission]{
		Type:    messaging.EighteenTerritoriesTwoTroops,
		Details: messaging.EighteenTerritoriesTwoTroopsMission{},
	}, nil
}

func (m *MissionControllerImpl) GetTwentyFourTerritoriesMission(
	_ ctx.GameContext,
	_ int64,
) (messaging.MissionState[messaging.TwentyFourTerritoriesMission], error) {
	return messaging.MissionState[messaging.TwentyFourTerritoriesMission]{
		Type:    messaging.TwentyFourTerritories,
		Details: messaging.TwentyFourTerritoriesMission{},
	}, nil
}
