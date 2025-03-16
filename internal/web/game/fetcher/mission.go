package fetcher

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
)

type MissionFetcher interface {
	Fetcher
}

type MissionFetcherImpl struct {
	missionService    mission.Service
	missionController controller.MissionController
}

var _ MissionFetcher = (*MissionFetcherImpl)(nil)

type MissionFetcherResult struct {
	fx.Out

	MissionFetcher MissionFetcher
	Fetcher        Fetcher `group:"private_fetchers"`
}

func NewMissionFetcher(
	missionService mission.Service,
	missionController controller.MissionController,
) MissionFetcherResult {
	res := &MissionFetcherImpl{
		missionService:    missionService,
		missionController: missionController,
	}

	return MissionFetcherResult{
		MissionFetcher: res,
		Fetcher:        res,
	}
}

func (f *MissionFetcherImpl) FetchState(ctx ctx.GameContext, stateChannel chan json.RawMessage) {
	ctx.Log().Debugw("fetching mission state")

	baseMission, err := f.missionService.GetBaseMission(ctx)
	if err != nil {
		ctx.Log().Errorw("failed to get base mission", "error", err)

		return
	}

	switch baseMission.Type {
	case sqlc.GameMissionTypeTWOCONTINENTS:
		FetchState(
			ctx,
			message.MissionState,
			getFetcherFunc(f.missionController.GetTwoContinentsMission, baseMission.ID),
			stateChannel)
	case sqlc.GameMissionTypeTWOCONTINENTSPLUSONE:
		FetchState(
			ctx,
			message.MissionState,
			getFetcherFunc(f.missionController.GetTwoContinentsPlusOneMission, baseMission.ID),
			stateChannel)
	case sqlc.GameMissionTypeELIMINATEPLAYER:
		FetchState(
			ctx,
			message.MissionState,
			getFetcherFunc(f.missionController.GetEliminatePlayerMission, baseMission.ID),
			stateChannel)
	case sqlc.GameMissionTypeEIGHTEENTERRITORIESTWOTROOPS:
		FetchState(
			ctx,
			message.MissionState,
			getFetcherFunc(
				f.missionController.GetEighteenTerritoriesTwoTroopsMission,
				baseMission.ID,
			),
			stateChannel,
		)
	case sqlc.GameMissionTypeTWENTYFOURTERRITORIES:
		FetchState(
			ctx,
			message.MissionState,
			getFetcherFunc(f.missionController.GetTwentyFourTerritoriesMission, baseMission.ID),
			stateChannel)
	default:
		ctx.Log().Errorf("unknown mission type: %v", baseMission.Type)
	}
}

func getFetcherFunc[T messaging.MissionDetails](
	fetcherFunc func(ctx.GameContext, int64) (messaging.MissionState[T], error),
	missionID int64,
) func(ctx.GameContext) (messaging.MissionState[T], error) {
	return func(cont ctx.GameContext) (messaging.MissionState[T], error) {
		return fetcherFunc(cont, missionID)
	}
}
