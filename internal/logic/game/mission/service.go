package mission

import (
	"errors"
	"fmt"
	"slices"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	"github.com/go-risk-it/go-risk-it/internal/rand"
)

type Service interface {
	CreateMissionsQ(
		ctx ctx.GameContext,
		querier db.Querier,
		players []sqlc.Player,
	) error
	IsMissionFulfilledQ(ctx ctx.GameContext, querier db.Querier) (bool, error)
}

type ServiceImpl struct {
	rng           rand.RNG
	boardService  board.Service
	regionService region.Service
}

var _ Service = (*ServiceImpl)(nil)

func New(
	rng rand.RNG,
	boardService board.Service,
	regionService region.Service,
) *ServiceImpl {
	return &ServiceImpl{
		rng:           rng,
		boardService:  boardService,
		regionService: regionService,
	}
}

func (s *ServiceImpl) CreateMissionsQ(
	ctx ctx.GameContext,
	querier db.Querier,
	players []sqlc.Player,
) error {
	ctx.Log().Infow("creating missions")

	missions := s.GetAvailableMissions(players)
	s.rng.Shuffle(len(missions), func(i, j int) {
		missions[i], missions[j] = missions[j], missions[i]
	})

	usedMissions := make([]bool, len(missions))

	for index := range players {
		mission, err := s.pickMission(players[index], missions, usedMissions)
		if err != nil {
			return fmt.Errorf("failed to pick mission: %w", err)
		}

		ctx.Log().Debugw("picked mission", "mission", mission)

		missionID, err := querier.InsertMission(ctx, sqlc.InsertMissionParams{
			PlayerID: players[index].ID,
			Type:     mission.Type(),
		})
		if err != nil {
			return fmt.Errorf("failed to insert missions: %w", err)
		}

		if err := mission.Persist(ctx, querier, missionID); err != nil {
			return fmt.Errorf("failed to persist mission specifics: %w", err)
		}
	}

	ctx.Log().Infow("created missions")

	return nil
}

func (s *ServiceImpl) pickMission(
	player sqlc.Player,
	missions []Mission,
	usedMissions []bool,
) (Mission, error) {
	for index := range missions {
		mission, isEliminatePlayer := missions[index].(*EliminatePlayerMission)

		isSuicidal := isEliminatePlayer && mission.TargetPlayerID == player.ID
		if !usedMissions[index] && !isSuicidal {
			usedMissions[index] = true

			return missions[index], nil
		}
	}

	return nil, errors.New("no missions left")
}

func (s *ServiceImpl) GetAvailableMissions(players []sqlc.Player) []Mission {
	missions := []Mission{
		&EighteenTerritoriesTwoTroopsMission{},
		&TwentyFourTerritoriesMission{},
		&TwoContinentsMission{
			Continent1: "north_america",
			Continent2: "africa",
		},
		&TwoContinentsMission{
			Continent1: "north_america",
			Continent2: "oceania",
		},
		&TwoContinentsMission{
			Continent1: "asia",
			Continent2: "south_america",
		},
		&TwoContinentsMission{
			Continent1: "asia",
			Continent2: "africa",
		},
		&TwoContinentsPlusOneMission{
			Continent1: "europe",
			Continent2: "south_america",
		},
		&TwoContinentsPlusOneMission{
			Continent1: "europe",
			Continent2: "oceania",
		},
	}

	eliminatePlayerMissions := make([]Mission, len(players))
	for i := range eliminatePlayerMissions {
		eliminatePlayerMissions[i] = &EliminatePlayerMission{
			TargetPlayerID: players[i].ID,
		}
	}

	missions = append(missions, eliminatePlayerMissions...)

	return missions
}

func (s *ServiceImpl) IsMissionFulfilledQ(ctx ctx.GameContext, querier db.Querier) (bool, error) {
	ctx.Log().Debugw("checking if mission is fulfilled")

	baseMission, err := querier.GetMission(ctx, sqlc.GetMissionParams{
		GameID: ctx.GameID(),
		UserID: ctx.UserID(),
	})
	if err != nil {
		return false, fmt.Errorf("failed to get mission: %w", err)
	}

	switch baseMission.Type {
	case sqlc.MissionTypeTWOCONTINENTS:
		return s.isTwoContinentsMissionFulfilled(ctx, querier, baseMission)
	case sqlc.MissionTypeTWOCONTINENTSPLUSONE:
		return s.isTwoContinentsPlusOneMissionFulfilled(ctx, querier, baseMission)
	case sqlc.MissionTypeEIGHTEENTERRITORIESTWOTROOPS:
		return s.isEighteenTerritoriesTwoTroopsMissionFulfilled(ctx, querier, baseMission)
	case sqlc.MissionTypeTWENTYFOURTERRITORIES:
		return s.isTwentyFourTerritoriesMissionFulfilled(ctx, querier, baseMission)
	case sqlc.MissionTypeELIMINATEPLAYER:
		return s.isEliminatePlayerMissionFulfilled(ctx, querier, baseMission)
	default:
		return false, fmt.Errorf("unknown mission type: %s", baseMission.Type)
	}
}

func continentEquals(cont string) func(continent *board.Continent) bool {
	return func(continent *board.Continent) bool {
		return continent.ExternalReference == cont
	}
}

func (s *ServiceImpl) isTwoContinentsMissionFulfilled(
	ctx ctx.GameContext,
	querier db.Querier,
	baseMission sqlc.Mission,
) (bool, error) {
	mission, err := querier.GetTwoContinentsMission(ctx, baseMission.ID)
	if err != nil {
		return false, fmt.Errorf("failed to get two continents mission: %w", err)
	}

	continents, err := s.boardService.GetContinentsControlledByPlayerQ(ctx, querier)
	if err != nil {
		return false, fmt.Errorf("failed to get continents controlled by player: %w", err)
	}

	return slices.ContainsFunc(continents, continentEquals(mission.Continent1)) &&
		slices.ContainsFunc(continents, continentEquals(mission.Continent2)), nil
}

func (s *ServiceImpl) isTwoContinentsPlusOneMissionFulfilled(
	ctx ctx.GameContext,
	querier db.Querier,
	baseMission sqlc.Mission,
) (bool, error) {
	mission, err := querier.GetTwoContinentsPlusOneMission(ctx, baseMission.ID)
	if err != nil {
		return false, fmt.Errorf("failed to get two continents plus one mission: %w", err)
	}

	continents, err := s.boardService.GetContinentsControlledByPlayerQ(ctx, querier)
	if err != nil {
		return false, fmt.Errorf("failed to get continents controlled by player: %w", err)
	}

	playerControlsTwoMandatoryContinents := slices.ContainsFunc(
		continents,
		continentEquals(mission.Continent1),
	) &&
		slices.ContainsFunc(continents, continentEquals(mission.Continent2))

	return playerControlsTwoMandatoryContinents && len(continents) > 2, nil
}

func (s *ServiceImpl) isEighteenTerritoriesTwoTroopsMissionFulfilled(
	ctx ctx.GameContext,
	querier db.Querier,
	_ sqlc.Mission,
) (bool, error) {
	regions, err := s.regionService.GetPlayerRegionsQ(ctx, querier)
	if err != nil {
		return false, fmt.Errorf("failed to get player regions: %w", err)
	}

	count := 0

	for _, region := range regions {
		if region.Troops > 1 {
			count++
		}
	}

	return count >= 18, nil
}

func (s *ServiceImpl) isTwentyFourTerritoriesMissionFulfilled(
	ctx ctx.GameContext,
	querier db.Querier,
	_ sqlc.Mission,
) (bool, error) {
	regions, err := s.regionService.GetPlayerRegionsQ(ctx, querier)
	if err != nil {
		return false, fmt.Errorf("failed to get player regions: %w", err)
	}

	return len(regions) >= 24, nil
}

func (s *ServiceImpl) isEliminatePlayerMissionFulfilled(
	ctx ctx.GameContext,
	querier db.Querier,
	baseMission sqlc.Mission,
) (bool, error) {
	mission, err := querier.GetEliminatePlayerMission(ctx, baseMission.ID)
	if err != nil {
		return false, fmt.Errorf("failed to get eliminate player mission: %w", err)
	}

	targetPlayerRegions, err := s.regionService.GetRegionsControlledByPlayerQ(
		ctx,
		querier,
		mission.TargetPlayerID,
	)
	if err != nil {
		return false, fmt.Errorf("failed to get player regions: %w", err)
	}

	return len(targetPlayerRegions) == 0, nil
}
