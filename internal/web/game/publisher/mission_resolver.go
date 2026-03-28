package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/game/converter"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
)

// BuildMissionResolver creates a MissionResolver closure that dispatches
// to the correct MissionController method based on mission type, wrapping
// the typed result into a json.RawMessage envelope.
func BuildMissionResolver(
	missionController *controller.MissionController,
) converter.MissionResolver {
	return func(
		c context.Context,
		missionType sqlc.GameMissionType,
		missionID int64,
	) (json.RawMessage, error) {
		gameCtx, ok := c.(ctx.GameContext)
		if !ok {
			return nil, errors.New("mission resolver requires GameContext")
		}

		return resolveMission(gameCtx, missionController, missionType, missionID)
	}
}

func resolveMission(
	gameCtx ctx.GameContext,
	missionCtrl *controller.MissionController,
	missionType sqlc.GameMissionType,
	missionID int64,
) (json.RawMessage, error) {
	switch missionType {
	case sqlc.GameMissionTypeTWOCONTINENTS:
		return fetchAndBuildMission(missionCtrl.GetTwoContinentsMission, gameCtx, missionID)
	case sqlc.GameMissionTypeTWOCONTINENTSPLUSONE:
		return fetchAndBuildMission(missionCtrl.GetTwoContinentsPlusOneMission, gameCtx, missionID)
	case sqlc.GameMissionTypeELIMINATEPLAYER:
		return fetchAndBuildMission(missionCtrl.GetEliminatePlayerMission, gameCtx, missionID)
	case sqlc.GameMissionTypeEIGHTEENTERRITORIESTWOTROOPS:
		return fetchAndBuildMission(
			missionCtrl.GetEighteenTerritoriesTwoTroopsMission,
			gameCtx,
			missionID,
		)
	case sqlc.GameMissionTypeTWENTYFOURTERRITORIES:
		return fetchAndBuildMission(
			missionCtrl.GetTwentyFourTerritoriesMission,
			gameCtx,
			missionID,
		)
	default:
		return nil, fmt.Errorf("unknown mission type: %s", missionType)
	}
}

func fetchAndBuildMission[T any](
	fetch func(ctx.GameContext, int64) (T, error),
	gameCtx ctx.GameContext,
	missionID int64,
) (json.RawMessage, error) {
	state, err := fetch(gameCtx, missionID)
	if err != nil {
		return nil, fmt.Errorf("fetching mission: %w", err)
	}

	return message.BuildMessage(message.MissionState, state)
}
