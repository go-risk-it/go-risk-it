package mission

import (
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission/checker"
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

	GetBaseMission(ctx ctx.GameContext) (sqlc.GameMission, error)
	GetTwoContinentsMission(
		ctx ctx.GameContext,
		missionID int64,
	) (TwoContinentsMission, error)
	GetTwoContinentsPlusOneMission(
		ctx ctx.GameContext,
		missionID int64,
	) (TwoContinentsPlusOneMission, error)
	GetEliminatePlayerMission(
		ctx ctx.GameContext,
		missionID int64,
	) (string, error)
}
type ServiceImpl struct {
	rng             rand.RNG
	querier         db.Querier
	checkerRegistry *checker.Registry
}

var _ Service = (*ServiceImpl)(nil)

func New(
	rng rand.RNG,
	querier db.Querier,
	checkerRegistry *checker.Registry,
) *ServiceImpl {
	return &ServiceImpl{
		rng:             rng,
		querier:         querier,
		checkerRegistry: checkerRegistry,
	}
}

func (s *ServiceImpl) GetBaseMission(ctx ctx.GameContext) (sqlc.GameMission, error) {
	baseMission, err := s.querier.GetMission(ctx, sqlc.GetMissionParams{
		GameID: ctx.GameID(),
		UserID: ctx.UserID(),
	})
	if err != nil {
		return sqlc.GameMission{}, fmt.Errorf("failed to get mission: %w", err)
	}

	return baseMission, nil
}

func (s *ServiceImpl) GetTwoContinentsMission(
	ctx ctx.GameContext,
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

func (s *ServiceImpl) GetTwoContinentsPlusOneMission(
	ctx ctx.GameContext,
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

func (s *ServiceImpl) GetEliminatePlayerMission(
	ctx ctx.GameContext,
	missionID int64,
) (string, error) {
	targetUser, err := s.querier.GetPlayerToEliminate(ctx, missionID)
	if err != nil {
		return "", fmt.Errorf("failed to get eliminate player mission: %w", err)
	}

	return targetUser, nil
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

		if err := mission.PersistQ(ctx, querier, missionID); err != nil {
			return fmt.Errorf("failed to persist mission specifics: %w", err)
		}
	}

	ctx.Log().Infow("created missions")

	return nil
}

func (s *ServiceImpl) pickMission(
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

func (s *ServiceImpl) GetAvailableMissions(players []sqlc.GamePlayer) []BaseMission {
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
	c, err := s.checkerRegistry.GetChecker(baseMission.Type)
	if err != nil {
		return false, err
	}

	return c.Check(ctx, querier, baseMission)
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
