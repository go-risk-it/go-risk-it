package consumers

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/consumers/converter"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
)

// BuildMissionResolver creates a MissionResolver closure that dispatches
// to the correct MissionController method based on mission type, returning
// the typed DTO for serialization at the dispatch boundary.
func BuildMissionResolver(
	missionController *MissionController,
) converter.MissionResolver {
	return func(
		c context.Context,
		missionType sqlc.GameMissionType,
		missionID int64,
	) (any, error) {
		gameCtx, ok := c.(ctx.GameContext)
		if !ok {
			return nil, errors.New("mission resolver requires GameContext")
		}

		return resolveMission(gameCtx, missionController, missionType, missionID)
	}
}

func resolveMission(
	gameCtx ctx.GameContext,
	missionCtrl *MissionController,
	missionType sqlc.GameMissionType,
	missionID int64,
) (any, error) {
	switch missionType {
	case sqlc.GameMissionTypeTWOCONTINENTS:
		return fetchMission(missionCtrl.GetTwoContinentsMission, gameCtx, missionID)
	case sqlc.GameMissionTypeTWOCONTINENTSPLUSONE:
		return fetchMission(missionCtrl.GetTwoContinentsPlusOneMission, gameCtx, missionID)
	case sqlc.GameMissionTypeELIMINATEPLAYER:
		return fetchMission(missionCtrl.GetEliminatePlayerMission, gameCtx, missionID)
	case sqlc.GameMissionTypeEIGHTEENTERRITORIESTWOTROOPS:
		return fetchMission(
			missionCtrl.GetEighteenTerritoriesTwoTroopsMission,
			gameCtx,
			missionID,
		)
	case sqlc.GameMissionTypeTWENTYFOURTERRITORIES:
		return fetchMission(
			missionCtrl.GetTwentyFourTerritoriesMission,
			gameCtx,
			missionID,
		)
	default:
		return nil, fmt.Errorf("unknown mission type: %s", missionType)
	}
}

func fetchMission[T any](
	fetch func(ctx.GameContext, int64) (T, error),
	gameCtx ctx.GameContext,
	missionID int64,
) (any, error) {
	state, err := fetch(gameCtx, missionID)
	if err != nil {
		return nil, fmt.Errorf("fetching mission: %w", err)
	}

	return state, nil
}
