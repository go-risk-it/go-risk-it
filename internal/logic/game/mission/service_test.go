package mission_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	board2 "github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/board"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/region"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/rand"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setup(t *testing.T) (
	*db.Querier,
	*board.Service,
	*region.Service,
	*mission.ServiceImpl,
) {
	t.Helper()
	querier := db.NewQuerier(t)
	boardService := board.NewService(t)
	regionService := region.NewService(t)
	rng := rand.NewRNG(t)

	service := mission.New(rng, boardService, regionService)

	return querier, boardService, regionService, service
}

func input() ctx.GameContext {
	gameID := int64(1)
	userID := "giovanni"

	userContext := ctx.WithUserID(
		ctx.WithLog(context.Background(), zap.NewExample().Sugar()),
		userID,
	)

	return ctx.WithGameID(userContext, gameID)
}

func TestServiceImpl_IsTwoContinentsMissionAccomplished(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name                 string
		controlledContinents []*board2.Continent
		missionContinent1    string
		missionContinent2    string
		expectedResult       bool
	}

	tests := []inputType{
		{
			"player does not control any continent",
			[]*board2.Continent{},
			"asia",
			"europe",
			false,
		},
		{
			"player controls one continent",
			[]*board2.Continent{
				{
					ExternalReference: "north_america",
					BonusTroops:       5,
				},
			},
			"asia",
			"europe",
			false,
		},
		{
			"one controlled but not the other",
			[]*board2.Continent{
				{
					ExternalReference: "north_america",
					BonusTroops:       5,
				},
				{
					ExternalReference: "africa",
					BonusTroops:       3,
				},
			},
			"north_america",
			"south_america",
			false,
		},
		{
			"both controlled",
			[]*board2.Continent{
				{
					ExternalReference: "north_america",
					BonusTroops:       5,
				},
				{
					ExternalReference: "africa",
					BonusTroops:       3,
				},
			},
			"north_america",
			"africa",
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, boardService, _, service := setup(t)
			ctx := input()

			baseMission := sqlc.Mission{
				ID:       1,
				PlayerID: 1,
				Type:     sqlc.MissionTypeTWOCONTINENTS,
			}

			querier.
				EXPECT().
				GetMission(ctx, sqlc.GetMissionParams{
					GameID: ctx.GameID(),
					UserID: ctx.UserID(),
				}).Return(baseMission, nil)

			twoContinentsMission := sqlc.TwoContinentsMission{
				MissionID:  baseMission.ID,
				Continent1: test.missionContinent1,
				Continent2: test.missionContinent2,
			}

			querier.
				EXPECT().
				GetTwoContinentsMission(ctx, baseMission.ID).
				Return(twoContinentsMission, nil)

			boardService.
				EXPECT().
				GetContinentsControlledByPlayerQ(ctx, querier).
				Return(test.controlledContinents, nil)

			result, err := service.IsMissionAccomplishedQ(ctx, querier)

			require.NoError(t, err)
			require.Equal(t, test.expectedResult, result)
		})
	}
}

func TestServiceImpl_IsTwoContinentsPlusOneMissionAccomplished(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name                 string
		controlledContinents []*board2.Continent
		missionContinent1    string
		missionContinent2    string
		expectedResult       bool
	}

	tests := []inputType{
		{
			"player does not control any continent",
			[]*board2.Continent{},
			"asia",
			"europe",
			false,
		},
		{
			"player controls one continent",
			[]*board2.Continent{
				{
					ExternalReference: "north_america",
					BonusTroops:       5,
				},
			},
			"asia",
			"europe",
			false,
		},
		{
			"one controlled but not the other",
			[]*board2.Continent{
				{
					ExternalReference: "north_america",
					BonusTroops:       5,
				},
				{
					ExternalReference: "africa",
					BonusTroops:       3,
				},
			},
			"north_america",
			"south_america",
			false,
		},
		{
			"both controlled, but no third continent",
			[]*board2.Continent{
				{
					ExternalReference: "north_america",
					BonusTroops:       5,
				},
				{
					ExternalReference: "africa",
					BonusTroops:       3,
				},
			},
			"north_america",
			"africa",
			false,
		},
		{
			"controls both continents and a third",
			[]*board2.Continent{
				{
					ExternalReference: "north_america",
					BonusTroops:       5,
				},
				{
					ExternalReference: "africa",
					BonusTroops:       3,
				},
				{
					ExternalReference: "south_america",
					BonusTroops:       2,
				},
			},
			"north_america",
			"africa",
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, boardService, _, service := setup(t)
			ctx := input()

			baseMission := sqlc.Mission{
				ID:       1,
				PlayerID: 1,
				Type:     sqlc.MissionTypeTWOCONTINENTSPLUSONE,
			}

			querier.
				EXPECT().
				GetMission(ctx, sqlc.GetMissionParams{
					GameID: ctx.GameID(),
					UserID: ctx.UserID(),
				}).Return(baseMission, nil)

			twoContinentsMission := sqlc.TwoContinentsPlusOneMission{
				MissionID:  baseMission.ID,
				Continent1: test.missionContinent1,
				Continent2: test.missionContinent2,
			}

			querier.
				EXPECT().
				GetTwoContinentsPlusOneMission(ctx, baseMission.ID).
				Return(twoContinentsMission, nil)

			boardService.
				EXPECT().
				GetContinentsControlledByPlayerQ(ctx, querier).
				Return(test.controlledContinents, nil)

			result, err := service.IsMissionAccomplishedQ(ctx, querier)

			require.NoError(t, err)
			require.Equal(t, test.expectedResult, result)
		})
	}
}

func TestServiceImpl_IsEliminatePlayerMissionAccomplished(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name                      string
		regionsControlledByTarget []sqlc.Region
		expectedResult            bool
	}

	tests := []inputType{
		{
			"target controls zero regions",
			[]sqlc.Region{},
			true,
		},
		{
			"player controls one continent",
			[]sqlc.Region{
				{
					ID:                1,
					ExternalReference: "quebec",
				},
			},
			false,
		},
		{
			"player controls two continents",
			[]sqlc.Region{
				{
					ID:                1,
					ExternalReference: "quebec",
				},
				{
					ID:                2,
					ExternalReference: "ontario",
				},
			},
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, _, regionService, service := setup(t)
			ctx := input()

			baseMission := sqlc.Mission{
				ID:       1,
				PlayerID: 1,
				Type:     sqlc.MissionTypeELIMINATEPLAYER,
			}

			querier.
				EXPECT().
				GetMission(ctx, sqlc.GetMissionParams{
					GameID: ctx.GameID(),
					UserID: ctx.UserID(),
				}).Return(baseMission, nil)

			eliminatePlayerMission := sqlc.EliminatePlayerMission{
				MissionID:      baseMission.ID,
				TargetPlayerID: 2,
			}

			querier.
				EXPECT().
				GetEliminatePlayerMission(ctx, baseMission.ID).
				Return(eliminatePlayerMission, nil)

			regionService.
				EXPECT().
				GetRegionsControlledByPlayerQ(ctx, querier, int64(2)).
				Return(test.regionsControlledByTarget, nil)

			result, err := service.IsMissionAccomplishedQ(ctx, querier)

			require.NoError(t, err)
			require.Equal(t, test.expectedResult, result)
		})
	}
}
