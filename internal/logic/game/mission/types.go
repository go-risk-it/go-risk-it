package mission

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
)

type Mission interface {
	Type() sqlc.GameMissionType
	Persist(ctx ctx.GameContext, querier db.Querier, missionID int64) error
}

type TwoContinentsMission struct {
	Continent1 string
	Continent2 string
}

type TwoContinentsPlusOneMission struct {
	Continent1 string
	Continent2 string
}

type EighteenTerritoriesTwoTroopsMission struct{}

type TwentyFourTerritoriesMission struct{}

type EliminatePlayerMission struct {
	TargetPlayerID int64
}

func (m *TwoContinentsMission) Type() sqlc.GameMissionType {
	return sqlc.GameMissionTypeTWOCONTINENTS
}

func (m *TwoContinentsPlusOneMission) Type() sqlc.GameMissionType {
	return sqlc.GameMissionTypeTWOCONTINENTSPLUSONE
}

func (m *EighteenTerritoriesTwoTroopsMission) Type() sqlc.GameMissionType {
	return sqlc.GameMissionTypeEIGHTEENTERRITORIESTWOTROOPS
}

func (m *TwentyFourTerritoriesMission) Type() sqlc.GameMissionType {
	return sqlc.GameMissionTypeTWENTYFOURTERRITORIES
}

func (m *EliminatePlayerMission) Type() sqlc.GameMissionType {
	return sqlc.GameMissionTypeELIMINATEPLAYER
}

func (m *TwoContinentsMission) Persist(
	ctx ctx.GameContext,
	querier db.Querier,
	missionID int64,
) error {
	return querier.InsertTwoContinentsMission(
		ctx,
		sqlc.InsertTwoContinentsMissionParams{
			MissionID:  missionID,
			Continent1: m.Continent1,
			Continent2: m.Continent2,
		})
}

func (m *TwoContinentsPlusOneMission) Persist(
	ctx ctx.GameContext,
	querier db.Querier,
	missionID int64,
) error {
	return querier.InsertTwoContinentsPlusOneMission(
		ctx,
		sqlc.InsertTwoContinentsPlusOneMissionParams{
			MissionID:  missionID,
			Continent1: m.Continent1,
			Continent2: m.Continent2,
		})
}

func (m *EliminatePlayerMission) Persist(
	ctx ctx.GameContext,
	querier db.Querier,
	missionID int64,
) error {
	return querier.InsertEliminatePlayerMission(
		ctx,
		sqlc.InsertEliminatePlayerMissionParams{
			MissionID:      missionID,
			TargetPlayerID: m.TargetPlayerID,
		})
}

func (m *EighteenTerritoriesTwoTroopsMission) Persist(
	_ ctx.GameContext,
	_ db.Querier,
	_ int64,
) error {
	return nil
}

func (m *TwentyFourTerritoriesMission) Persist(
	_ ctx.GameContext,
	_ db.Querier,
	_ int64,
) error {
	return nil
}

var (
	_ Mission = (*TwoContinentsMission)(nil)
	_ Mission = (*TwoContinentsPlusOneMission)(nil)
	_ Mission = (*EighteenTerritoriesTwoTroopsMission)(nil)
	_ Mission = (*TwentyFourTerritoriesMission)(nil)
	_ Mission = (*EliminatePlayerMission)(nil)
)
