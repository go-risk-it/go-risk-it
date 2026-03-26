package fetcher

import (
	"encoding/json"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission"
	sharedfetcher "github.com/go-risk-it/go-risk-it/internal/web/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
)

type missionFetcher struct {
	missionService    mission.Service
	missionController *controller.MissionController
}

type MissionFetcherResult struct {
	fx.Out

	Fetcher Fetcher `group:"private_fetchers"`
}

func NewMissionFetcher(
	missionService mission.Service,
	missionController *controller.MissionController,
) MissionFetcherResult {
	return MissionFetcherResult{
		Fetcher: &missionFetcher{
			missionService:    missionService,
			missionController: missionController,
		},
	}
}

func (f *missionFetcher) FetchState(ctx ctx.GameContext, stateChannel chan json.RawMessage) {
	slog.DebugContext(ctx, "fetching mission state")

	baseMission, err := f.missionService.GetBaseMission(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get base mission", "error", err)

		return
	}

	switch baseMission.Type {
	case sqlc.GameMissionTypeTWOCONTINENTS:
		sharedfetcher.FetchState(
			ctx,
			message.MissionState,
			getFetcherFunc(f.missionController.GetTwoContinentsMission, baseMission.ID),
			stateChannel)
	case sqlc.GameMissionTypeTWOCONTINENTSPLUSONE:
		sharedfetcher.FetchState(
			ctx,
			message.MissionState,
			getFetcherFunc(f.missionController.GetTwoContinentsPlusOneMission, baseMission.ID),
			stateChannel)
	case sqlc.GameMissionTypeELIMINATEPLAYER:
		sharedfetcher.FetchState(
			ctx,
			message.MissionState,
			getFetcherFunc(f.missionController.GetEliminatePlayerMission, baseMission.ID),
			stateChannel)
	case sqlc.GameMissionTypeEIGHTEENTERRITORIESTWOTROOPS:
		sharedfetcher.FetchState(
			ctx,
			message.MissionState,
			getFetcherFunc(
				f.missionController.GetEighteenTerritoriesTwoTroopsMission,
				baseMission.ID,
			),
			stateChannel,
		)
	case sqlc.GameMissionTypeTWENTYFOURTERRITORIES:
		sharedfetcher.FetchState(
			ctx,
			message.MissionState,
			getFetcherFunc(f.missionController.GetTwentyFourTerritoriesMission, baseMission.ID),
			stateChannel)
	default:
		slog.ErrorContext(ctx, "unknown mission type", "type", baseMission.Type)
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
