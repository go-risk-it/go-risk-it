package mission

import (
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/rand"
)

type Service interface {
	CreateMissionsQ(
		ctx ctx.GameContext,
		querier db.Querier,
		players []sqlc.Player,
	) error
}

type ServiceImpl struct {
	rng rand.RNG
}

var _ Service = (*ServiceImpl)(nil)

func New(
	rng rand.RNG,
) *ServiceImpl {
	return &ServiceImpl{
		rng: rng,
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

		if err := mission.PersistSpecifics(ctx, querier, missionID); err != nil {
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
