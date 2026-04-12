package mission

import (
	"errors"
	"fmt"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/mission/checker"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/rand"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateMissions(
		ctx gamectx.GameContext,
		querier db.Querier,
		players []sqlc.GamePlayer,
	) error
	IsMissionAccomplished(ctx gamectx.GameContext, querier db.Querier) (bool, error)
	ReassignMissions(ctx gamectx.GameContext, querier db.Querier, eliminatedPlayerID int64) error

	GetTwoContinentsMission(
		ctx gamectx.GameContext,
		missionID int64,
	) (TwoContinentsMission, error)
	GetTwoContinentsPlusOneMission(
		ctx gamectx.GameContext,
		missionID int64,
	) (TwoContinentsPlusOneMission, error)
	GetEliminatePlayerMission(
		ctx gamectx.GameContext,
		missionID int64,
	) (string, error)
}

type service struct {
	rng             rand.RNG
	querier         db.Querier
	checkerRegistry *checker.Registry
}

var _ Service = (*service)(nil)

func New(
	rng rand.RNG,
	querier db.Querier,
	checkerRegistry *checker.Registry,
) Service {
	return &service{
		rng:             rng,
		querier:         querier,
		checkerRegistry: checkerRegistry,
	}
}

func (s *service) GetTwoContinentsMission(
	ctx gamectx.GameContext,
	missionID int64,
) (TwoContinentsMission, error) {
	mission, err := s.querier.GetTwoContinentsMission(ctx, missionID)
	if err != nil {
		return TwoContinentsMission{}, fmt.Errorf("failed to get two continents mission: %w", err)
	}

	return TwoContinentsMission{
		Continent1: mission.Continent1,
		Continent2: mission.Continent2,
	}, nil
}

func (s *service) GetTwoContinentsPlusOneMission(
	ctx gamectx.GameContext,
	missionID int64,
) (TwoContinentsPlusOneMission, error) {
	mission, err := s.querier.GetTwoContinentsPlusOneMission(ctx, missionID)
	if err != nil {
		return TwoContinentsPlusOneMission{}, fmt.Errorf(
			"failed to get two continents plus one mission: %w",
			err,
		)
	}

	return TwoContinentsPlusOneMission{
		Continent1: mission.Continent1,
		Continent2: mission.Continent2,
	}, nil
}

func (s *service) GetEliminatePlayerMission(
	ctx gamectx.GameContext,
	missionID int64,
) (string, error) {
	targetUser, err := s.querier.GetPlayerToEliminate(ctx, missionID)
	if err != nil {
		return "", fmt.Errorf("failed to get eliminate player mission: %w", err)
	}

	return targetUser, nil
}

func (s *service) CreateMissions(
	ctx gamectx.GameContext,
	querier db.Querier,
	players []sqlc.GamePlayer,
) error {
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

	return nil
}

func (s *service) pickMission(
	player sqlc.GamePlayer,
	missions []BaseMission,
	usedMissions []bool,
) (BaseMission, error) {
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

func (s *service) GetAvailableMissions(players []sqlc.GamePlayer) []BaseMission {
	missions := make([]BaseMission, 0, 8+len(players))
	missions = append(missions, []BaseMission{
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
	}...)

	eliminatePlayerMissions := make([]BaseMission, len(players))
	for i := range eliminatePlayerMissions {
		eliminatePlayerMissions[i] = &EliminatePlayerMission{
			TargetPlayerID: players[i].ID,
		}
	}

	missions = append(missions, eliminatePlayerMissions...)

	return missions
}

func (s *service) IsMissionAccomplished(
	ctx gamectx.GameContext,
	querier db.Querier,
) (bool, error) {
	return observe.Span(
		ctx,
		"game.move.check_mission",
		func(ctx gamectx.GameContext) (bool, error) {
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
		},
	)
}

func (s *service) isMissionAccomplished(
	ctx gamectx.GameContext,
	querier db.Querier,
	baseMission sqlc.GameMission,
) (bool, error) {
	c, err := s.checkerRegistry.GetChecker(baseMission.Type)
	if err != nil {
		return false, err
	}

	return c.Check(ctx, querier, baseMission)
}

func (s *service) ReassignMissions(
	ctx gamectx.GameContext,
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

	return nil
}
