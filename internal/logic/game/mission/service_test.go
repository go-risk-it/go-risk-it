package mission_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	board2 "github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission/checker"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/game/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/kernel/rand"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/board"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/region"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func setup(t *testing.T) (
	*db.Querier,
	*board.Service,
	*region.Service,
	mission.Service,
) {
	t.Helper()
	querier := db.NewQuerier(t)
	boardService := board.NewService(t)
	regionService := region.NewService(t)
	rng := rand.NewRNG(t)

	registry, err := checker.NewRegistry([]checker.MissionChecker{
		checker.NewTwoContinentsChecker(boardService),
		checker.NewTwoContinentsPlusOneChecker(boardService),
		checker.NewEighteenTerritoriesChecker(regionService),
		checker.NewTwentyFourTerritoriesChecker(regionService),
		checker.NewEliminatePlayerChecker(regionService),
	})
	require.NoError(t, err)

	service := mission.New(rng, querier, registry)

	return querier, boardService, regionService, service
}

func input() ctx.GameContext {
	gameID := int64(1)
	userID := "giovanni"

	userContext := ctx.WithUserID(
		ctx.WithSpan(context.Background(), noop.Span{}),
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

			baseMission := sqlc.GameMission{
				ID:       1,
				PlayerID: 1,
				Type:     sqlc.GameMissionTypeTWOCONTINENTS,
			}

			querier.
				EXPECT().
				GetMission(mock.Anything, sqlc.GetMissionParams{
					GameID: ctx.GameID(),
					UserID: ctx.UserID(),
				}).Return(baseMission, nil)

			twoContinentsMission := sqlc.GameTwoContinentsMission{
				MissionID:  baseMission.ID,
				Continent1: test.missionContinent1,
				Continent2: test.missionContinent2,
			}

			querier.
				EXPECT().
				GetTwoContinentsMission(mock.Anything, baseMission.ID).
				Return(twoContinentsMission, nil)

			boardService.
				EXPECT().
				GetContinentsControlledByPlayer(mock.Anything, querier, int64(1)).
				Return(test.controlledContinents, nil)

			if test.expectedResult {
				querier.
					EXPECT().
					AssignGameWinner(mock.Anything, sqlc.AssignGameWinnerParams{
						WinnerPlayerID: pgtype.Int8{
							Int64: 1,
							Valid: true,
						},
						GameID: ctx.GameID(),
					}).
					Return(nil)
			}

			result, err := service.IsMissionAccomplished(ctx, querier)

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

			baseMission := sqlc.GameMission{
				ID:       1,
				PlayerID: 1,
				Type:     sqlc.GameMissionTypeTWOCONTINENTSPLUSONE,
			}

			querier.
				EXPECT().
				GetMission(mock.Anything, sqlc.GetMissionParams{
					GameID: ctx.GameID(),
					UserID: ctx.UserID(),
				}).Return(baseMission, nil)

			twoContinentsMission := sqlc.GameTwoContinentsPlusOneMission{
				MissionID:  baseMission.ID,
				Continent1: test.missionContinent1,
				Continent2: test.missionContinent2,
			}

			querier.
				EXPECT().
				GetTwoContinentsPlusOneMission(mock.Anything, baseMission.ID).
				Return(twoContinentsMission, nil)

			boardService.
				EXPECT().
				GetContinentsControlledByPlayer(mock.Anything, querier, int64(1)).
				Return(test.controlledContinents, nil)

			if test.expectedResult {
				querier.
					EXPECT().
					AssignGameWinner(mock.Anything, sqlc.AssignGameWinnerParams{
						WinnerPlayerID: pgtype.Int8{
							Int64: 1,
							Valid: true,
						},
						GameID: ctx.GameID(),
					}).
					Return(nil)
			}

			result, err := service.IsMissionAccomplished(ctx, querier)

			require.NoError(t, err)
			require.Equal(t, test.expectedResult, result)
		})
	}
}

func TestServiceImpl_IsEighteenTerritoriesTwoTroopsMissionAccomplished(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name           string
		playerRegions  []sqlc.GetRegionsByGameRow
		expectedResult bool
	}

	tests := []inputType{
		{
			"17 regions with 2 troops is not enough",
			func() []sqlc.GetRegionsByGameRow {
				regions := make([]sqlc.GetRegionsByGameRow, 17)
				for i := range regions {
					regions[i] = sqlc.GetRegionsByGameRow{
						ID:                int64(i + 1),
						ExternalReference: fmt.Sprintf("region_%d", i+1),
						Troops:            2,
						UserID:            "giovanni",
					}
				}

				return regions
			}(),
			false,
		},
		{
			"18 regions with 2 troops is enough",
			func() []sqlc.GetRegionsByGameRow {
				regions := make([]sqlc.GetRegionsByGameRow, 18)
				for i := range regions {
					regions[i] = sqlc.GetRegionsByGameRow{
						ID:                int64(i + 1),
						ExternalReference: fmt.Sprintf("region_%d", i+1),
						Troops:            2,
						UserID:            "giovanni",
					}
				}

				return regions
			}(),
			true,
		},
		{
			"18 regions but only 17 with 2 troops is not enough",
			func() []sqlc.GetRegionsByGameRow {
				regions := make([]sqlc.GetRegionsByGameRow, 18)
				for i := range regions {
					regions[i] = sqlc.GetRegionsByGameRow{
						ID:                int64(i + 1),
						ExternalReference: fmt.Sprintf("region_%d", i+1),
						Troops:            2,
						UserID:            "giovanni",
					}
				}
				regions[0].Troops = 1

				return regions
			}(),
			false,
		},
		{
			"19 regions but only 18 with 2 troops is enough",
			func() []sqlc.GetRegionsByGameRow {
				regions := make([]sqlc.GetRegionsByGameRow, 19)
				for i := range regions {
					regions[i] = sqlc.GetRegionsByGameRow{
						ID:                int64(i + 1),
						ExternalReference: fmt.Sprintf("region_%d", i+1),
						Troops:            2,
						UserID:            "giovanni",
					}
				}
				regions[0].Troops = 1

				return regions
			}(),
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, _, regionService, service := setup(t)
			ctx := input()

			baseMission := sqlc.GameMission{
				ID:       1,
				PlayerID: 1,
				Type:     sqlc.GameMissionTypeEIGHTEENTERRITORIESTWOTROOPS,
			}

			querier.
				EXPECT().
				GetMission(mock.Anything, sqlc.GetMissionParams{
					GameID: ctx.GameID(),
					UserID: ctx.UserID(),
				}).Return(baseMission, nil)

			regionService.
				EXPECT().
				GetPlayerRegions(mock.Anything, querier).
				Return(test.playerRegions, nil)

			if test.expectedResult {
				querier.
					EXPECT().
					AssignGameWinner(mock.Anything, sqlc.AssignGameWinnerParams{
						WinnerPlayerID: pgtype.Int8{
							Int64: 1,
							Valid: true,
						},
						GameID: ctx.GameID(),
					}).
					Return(nil)
			}

			result, err := service.IsMissionAccomplished(ctx, querier)

			require.NoError(t, err)
			require.Equal(t, test.expectedResult, result)
		})
	}
}

func TestServiceImpl_IsTwentyFourTerritoriesMissionAccomplished(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name           string
		playerRegions  []sqlc.GetRegionsByGameRow
		expectedResult bool
	}

	tests := []inputType{
		{
			"23 regions is not enough",
			func() []sqlc.GetRegionsByGameRow {
				regions := make([]sqlc.GetRegionsByGameRow, 23)
				for i := range regions {
					regions[i] = sqlc.GetRegionsByGameRow{
						ID:                int64(i + 1),
						ExternalReference: fmt.Sprintf("region_%d", i+1),
						Troops:            1,
						UserID:            "giovanni",
					}
				}

				return regions
			}(),
			false,
		},
		{
			"24 regions is enough",
			func() []sqlc.GetRegionsByGameRow {
				regions := make([]sqlc.GetRegionsByGameRow, 24)
				for i := range regions {
					regions[i] = sqlc.GetRegionsByGameRow{
						ID:                int64(i + 1),
						ExternalReference: fmt.Sprintf("region_%d", i+1),
						Troops:            1,
						UserID:            "giovanni",
					}
				}

				return regions
			}(),
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, _, regionService, service := setup(t)
			ctx := input()

			baseMission := sqlc.GameMission{
				ID:       1,
				PlayerID: 1,
				Type:     sqlc.GameMissionTypeTWENTYFOURTERRITORIES,
			}

			querier.
				EXPECT().
				GetMission(mock.Anything, sqlc.GetMissionParams{
					GameID: ctx.GameID(),
					UserID: ctx.UserID(),
				}).Return(baseMission, nil)

			regionService.
				EXPECT().
				GetPlayerRegions(mock.Anything, querier).
				Return(test.playerRegions, nil)

			if test.expectedResult {
				querier.
					EXPECT().
					AssignGameWinner(mock.Anything, sqlc.AssignGameWinnerParams{
						WinnerPlayerID: pgtype.Int8{
							Int64: 1,
							Valid: true,
						},
						GameID: ctx.GameID(),
					}).
					Return(nil)
			}

			result, err := service.IsMissionAccomplished(ctx, querier)

			require.NoError(t, err)
			require.Equal(t, test.expectedResult, result)
		})
	}
}

func TestServiceImpl_IsEliminatePlayerMissionAccomplished(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name                      string
		regionsControlledByTarget []sqlc.GameRegion
		expectedResult            bool
	}

	tests := []inputType{
		{
			"target controls zero regions",
			[]sqlc.GameRegion{},
			true,
		},
		{
			"player controls one continent",
			[]sqlc.GameRegion{
				{
					ID:                1,
					ExternalReference: "quebec",
				},
			},
			false,
		},
		{
			"player controls two continents",
			[]sqlc.GameRegion{
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

			baseMission := sqlc.GameMission{
				ID:       1,
				PlayerID: 1,
				Type:     sqlc.GameMissionTypeELIMINATEPLAYER,
			}

			querier.
				EXPECT().
				GetMission(mock.Anything, sqlc.GetMissionParams{
					GameID: ctx.GameID(),
					UserID: ctx.UserID(),
				}).Return(baseMission, nil)

			eliminatePlayerMission := sqlc.GameEliminatePlayerMission{
				MissionID:      baseMission.ID,
				TargetPlayerID: 2,
			}

			querier.
				EXPECT().
				GetEliminatePlayerMission(mock.Anything, baseMission.ID).
				Return(eliminatePlayerMission, nil)

			regionService.
				EXPECT().
				GetRegionsControlledByPlayer(mock.Anything, querier, int64(2)).
				Return(test.regionsControlledByTarget, nil)

			if test.expectedResult {
				querier.
					EXPECT().
					AssignGameWinner(mock.Anything, sqlc.AssignGameWinnerParams{
						WinnerPlayerID: pgtype.Int8{
							Int64: 1,
							Valid: true,
						},
						GameID: ctx.GameID(),
					}).
					Return(nil)
			}

			result, err := service.IsMissionAccomplished(ctx, querier)

			require.NoError(t, err)
			require.Equal(t, test.expectedResult, result)
		})
	}
}
