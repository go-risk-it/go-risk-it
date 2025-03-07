package mission

import (
	"errors"
	"fmt"
	"slices"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	"github.com/go-risk-it/go-risk-it/internal/rand"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateMissionsQ(
		ctx ctx.GameContext,
		querier db.Querier,
		players []sqlc.GamePlayer,
	) error
	IsMissionAccomplishedQ(ctx ctx.GameContext, querier db.Querier) (bool, error)
	ReassignMissionsQ(ctx ctx.GameContext, querier db.Querier, eliminatedPlayerID int64) error
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
	players []sqlc.GamePlayer,
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
	player sqlc.GamePlayer,
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

func (s *ServiceImpl) GetAvailableMissions(players []sqlc.GamePlayer) []Mission {
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

func (s *ServiceImpl) IsMissionAccomplishedQ(
	ctx ctx.GameContext,
	querier db.Querier,
) (bool, error) {
	ctx.Log().Debugw("checking if mission is accomplished")

	baseMission, err := querier.GetMission(ctx, sqlc.GetMissionParams{
		GameID: ctx.GameID(),
		UserID: ctx.UserID(),
	})
	if err != nil {
		return false, fmt.Errorf("failed to get mission: %w", err)
	}

	isMissionAccomplished, err := s.isMissionAccomplished(ctx, querier, baseMission)
	if err != nil {
		return false, fmt.Errorf("failed to check if mission is accomplished: %w", err)
	}

	if isMissionAccomplished {
		ctx.Log().Infow("mission is accomplished, assigning winner")

		if err := querier.AssignGameWinner(ctx, sqlc.AssignGameWinnerParams{
			WinnerPlayerID: pgtype.Int8{
				Int64: baseMission.PlayerID,
				Valid: true,
			},
			GameID: ctx.GameID(),
		}); err != nil {
			return false, fmt.Errorf("failed to assign game winner: %w", err)
		}

		return true, nil
	}

	return false, nil
}

func (s *ServiceImpl) isMissionAccomplished(
	ctx ctx.GameContext,
	querier db.Querier,
	baseMission sqlc.GameMission,
) (bool, error) {
	switch baseMission.Type {
	case sqlc.GameMissionTypeTWOCONTINENTS:
		return s.isTwoContinentsMissionAccomplished(ctx, querier, baseMission)
	case sqlc.GameMissionTypeTWOCONTINENTSPLUSONE:
		return s.isTwoContinentsPlusOneMissionAccomplished(ctx, querier, baseMission)
	case sqlc.GameMissionTypeEIGHTEENTERRITORIESTWOTROOPS:
		return s.isEighteenTerritoriesTwoTroopsMissionAccomplished(ctx, querier, baseMission)
	case sqlc.GameMissionTypeTWENTYFOURTERRITORIES:
		return s.isTwentyFourTerritoriesMissionAccomplished(ctx, querier, baseMission)
	case sqlc.GameMissionTypeELIMINATEPLAYER:
		return s.isEliminatePlayerMissionAccomplished(ctx, querier, baseMission)
	default:
		return false, fmt.Errorf("unknown mission type: %s", baseMission.Type)
	}
}

func continentEquals(cont string) func(continent *board.Continent) bool {
	return func(continent *board.Continent) bool {
		return continent.ExternalReference == cont
	}
}

func (s *ServiceImpl) isTwoContinentsMissionAccomplished(
	ctx ctx.GameContext,
	querier db.Querier,
	baseMission sqlc.GameMission,
) (bool, error) {
	mission, err := querier.GetTwoContinentsMission(ctx, baseMission.ID)
	if err != nil {
		return false, fmt.Errorf("failed to get two continents mission: %w", err)
	}

	continents, err := s.boardService.GetContinentsControlledByPlayerQ(
		ctx,
		querier,
		baseMission.PlayerID,
	)
	if err != nil {
		return false, fmt.Errorf("failed to get continents controlled by player: %w", err)
	}

	return slices.ContainsFunc(continents, continentEquals(mission.Continent1)) &&
		slices.ContainsFunc(continents, continentEquals(mission.Continent2)), nil
}

func (s *ServiceImpl) isTwoContinentsPlusOneMissionAccomplished(
	ctx ctx.GameContext,
	querier db.Querier,
	baseMission sqlc.GameMission,
) (bool, error) {
	mission, err := querier.GetTwoContinentsPlusOneMission(ctx, baseMission.ID)
	if err != nil {
		return false, fmt.Errorf("failed to get two continents plus one mission: %w", err)
	}

	continents, err := s.boardService.GetContinentsControlledByPlayerQ(
		ctx,
		querier,
		baseMission.PlayerID,
	)
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

func (s *ServiceImpl) isEighteenTerritoriesTwoTroopsMissionAccomplished(
	ctx ctx.GameContext,
	querier db.Querier,
	_ sqlc.GameMission,
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

func (s *ServiceImpl) isTwentyFourTerritoriesMissionAccomplished(
	ctx ctx.GameContext,
	querier db.Querier,
	_ sqlc.GameMission,
) (bool, error) {
	regions, err := s.regionService.GetPlayerRegionsQ(ctx, querier)
	if err != nil {
		return false, fmt.Errorf("failed to get player regions: %w", err)
	}

	return len(regions) >= 24, nil
}

func (s *ServiceImpl) isEliminatePlayerMissionAccomplished(
	ctx ctx.GameContext,
	querier db.Querier,
	baseMission sqlc.GameMission,
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

func (s *ServiceImpl) ReassignMissionsQ(
	ctx ctx.GameContext,
	querier db.Querier,
	eliminatedPlayerID int64,
) error {
	if err := querier.ReassignMissions(ctx, sqlc.ReassignMissionsParams{
		GameID:             ctx.GameID(),
		UserID:             ctx.UserID(),
		EliminatedPlayerID: eliminatedPlayerID,
	}); err != nil {
		return fmt.Errorf("failed to reassign missions: %w", err)
	}

	if err := querier.DeleteSpuriousEliminatePlayerMissions(ctx, ctx.GameID()); err != nil {
		return fmt.Errorf("failed to delete spurious missions: %w", err)
	}

	ctx.Log().Infow("reassigned missions")

	return nil
}
