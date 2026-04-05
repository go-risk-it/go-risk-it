package snapshot

import (
	"errors"
	"fmt"

	game "github.com/go-risk-it/go-risk-it/internal/game/api"
	apisnapshot "github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
)

// MissionQuerier is the narrow interface the Reader needs for resolving mission
// details. It is a consumer-local projection of mission.Service — the reader
// does not need CreateMissions, IsMissionAccomplished, or ReassignMissions.
//
// The return types are local value types that mirror mission package types
// without importing game/logic/mission (which game/snapshot must not import per
// arch rule 23). FX wires the real mission.Service through an adapter function.
type MissionQuerier interface {
	GetTwoContinentsMission(
		ctx gamectx.GameContext,
		missionID int64,
	) (TwoContinentsResult, error)
	GetTwoContinentsPlusOneMission(
		ctx gamectx.GameContext,
		missionID int64,
	) (TwoContinentsPlusOneResult, error)
	GetEliminatePlayerMission(
		ctx gamectx.GameContext,
		missionID int64,
	) (string, error)
}

// TwoContinentsResult holds the two continent names for a two-continents
// mission. Mirrors mission.TwoContinentsMission without importing the logic
// package.
type TwoContinentsResult struct {
	Continent1 string
	Continent2 string
}

// TwoContinentsPlusOneResult holds the two continent names for a
// two-continents-plus-one mission. Mirrors mission.TwoContinentsPlusOneMission
// without importing the logic package.
type TwoContinentsPlusOneResult struct {
	Continent1 string
	Continent2 string
}

// Reader is the public interface for SnapshotReader implementations in this
// package. It is a package-local alias of the port interface, used to decouple
// the constructor return type from the api package at the FX wiring boundary.
type Reader = game.SnapshotReader

// reader adapts the sqlc-typed snapshot.Service and MissionQuerier into the
// clean api/snapshot types. It is the single implementation of
// game.SnapshotReader in the system.
type reader struct {
	snapshotService Service
	missionQuerier  MissionQuerier
}

var _ game.SnapshotReader = (*reader)(nil)

// NewReader creates a SnapshotReader that maps sqlc types to clean snapshot
// types by delegating to the existing Service and MissionQuerier.
func NewReader(snapshotService Service, missionQuerier MissionQuerier) Reader {
	return &reader{
		snapshotService: snapshotService,
		missionQuerier:  missionQuerier,
	}
}

// ReaderFactory creates a Reader bound to a caller-supplied transactional
// querier. The shared MissionQuerier is captured at construction time.
// This enables the orchestrator to build a Reader inside a transaction
// so that snapshot reads are consistent with the move that just committed.
type ReaderFactory = func(querier db.Querier) Reader

// NewReaderFactory returns a ReaderFactory that closes over the given
// MissionQuerier. Each call to the returned function creates a fresh
// Reader backed by the provided transactional querier.
func NewReaderFactory(missionQuerier MissionQuerier) ReaderFactory {
	return func(querier db.Querier) Reader {
		return NewReader(NewService(querier), missionQuerier)
	}
}

func (r *reader) GetPublicSnapshot(
	ctx gamectx.GameContext,
) (*apisnapshot.GameSnapshot, error) {
	return observe.Span(
		ctx,
		"snapshot.reader.get_public",
		func(ctx gamectx.GameContext) (*apisnapshot.GameSnapshot, error) {
			snap, err := r.snapshotService.GetPublicSnapshot(ctx)
			if err != nil {
				return nil, fmt.Errorf("getting public snapshot: %w", err)
			}

			return mapPublicSnapshot(snap)
		},
	)
}

func (r *reader) GetAllPrivateSnapshots(
	ctx gamectx.GameContext,
) (map[string]*apisnapshot.PlayerPrivate, error) {
	return observe.Span(
		ctx,
		"snapshot.reader.get_private",
		func(ctx gamectx.GameContext) (map[string]*apisnapshot.PlayerPrivate, error) {
			snapshots, err := r.snapshotService.GetPrivateSnapshotsByUser(ctx)
			if err != nil {
				return nil, fmt.Errorf("getting private snapshots: %w", err)
			}

			result := make(map[string]*apisnapshot.PlayerPrivate, len(snapshots))
			for userID, snap := range snapshots {
				private, err := r.mapPrivateSnapshot(ctx, snap)
				if err != nil {
					return nil, fmt.Errorf("resolving mission for user %q: %w", userID, err)
				}

				result[userID] = private
			}

			return result, nil
		},
	)
}

// ---------------------------------------------------------------------------
// Public snapshot mapping
// ---------------------------------------------------------------------------

func mapPublicSnapshot(snap *PublicSnapshot) (*apisnapshot.GameSnapshot, error) {
	phase, err := mapPhase(snap.Phase)
	if err != nil {
		return nil, fmt.Errorf("mapping phase: %w", err)
	}

	winnerUserID := ""
	if snap.Game.WinnerUserID.Valid {
		winnerUserID = snap.Game.WinnerUserID.String
	}

	return &apisnapshot.GameSnapshot{
		Game: apisnapshot.GameMeta{
			ID:           snap.Game.ID,
			Turn:         snap.Game.Turn,
			WinnerUserID: winnerUserID,
		},
		Phase:   phase,
		Regions: mapRegions(snap.Board),
		Players: mapPlayers(snap.Players),
	}, nil
}

func mapPhase(phase PhaseState) (apisnapshot.Phase, error) {
	phaseType, err := mapPhaseType(phase.Type)
	if err != nil {
		return apisnapshot.Phase{}, err
	}

	state, err := mapPhaseState(phase)
	if err != nil {
		return apisnapshot.Phase{}, err
	}

	return apisnapshot.Phase{
		Type:  phaseType,
		State: state,
	}, nil
}

func mapPhaseType(t sqlc.GamePhaseType) (apisnapshot.PhaseType, error) {
	switch t {
	case sqlc.GamePhaseTypeDEPLOY:
		return apisnapshot.PhaseDeploy, nil
	case sqlc.GamePhaseTypeATTACK:
		return apisnapshot.PhaseAttack, nil
	case sqlc.GamePhaseTypeCONQUER:
		return apisnapshot.PhaseConquer, nil
	case sqlc.GamePhaseTypeREINFORCE:
		return apisnapshot.PhaseReinforce, nil
	case sqlc.GamePhaseTypeCARDS:
		return apisnapshot.PhaseCards, nil
	default:
		return "", fmt.Errorf("unknown phase type: %s", t)
	}
}

func mapPhaseState(phase PhaseState) (apisnapshot.PhaseState, error) {
	switch phase.Type {
	case sqlc.GamePhaseTypeDEPLOY:
		if phase.DeployState == nil {
			return nil, errors.New("deploy phase state is nil")
		}

		return apisnapshot.DeployPhaseState{
			DeployableTroops: phase.DeployState.DeployableTroops,
		}, nil
	case sqlc.GamePhaseTypeCONQUER:
		if phase.ConquerState == nil {
			return nil, errors.New("conquer phase state is nil")
		}

		return apisnapshot.ConquerPhaseState{
			AttackingRegionID: phase.ConquerState.SourceRegion,
			DefendingRegionID: phase.ConquerState.TargetRegion,
			MinTroopsToMove:   phase.ConquerState.MinimumTroops,
		}, nil
	case sqlc.GamePhaseTypeATTACK, sqlc.GamePhaseTypeREINFORCE, sqlc.GamePhaseTypeCARDS:
		return apisnapshot.EmptyPhaseState{}, nil
	default:
		return nil, fmt.Errorf("unknown phase type: %s", phase.Type)
	}
}

func mapRegions(regions []sqlc.GetRegionsByGameRow) []apisnapshot.RegionState {
	result := make([]apisnapshot.RegionState, len(regions))
	for i, r := range regions {
		result[i] = apisnapshot.RegionState{
			ID:      r.ExternalReference,
			OwnerID: r.UserID,
			Troops:  r.Troops,
		}
	}

	return result
}

func mapPlayers(players []sqlc.GetPlayersStateRow) []apisnapshot.PlayerState {
	result := make([]apisnapshot.PlayerState, len(players))
	for i, p := range players {
		result[i] = apisnapshot.PlayerState{
			UserID:    p.UserID,
			Name:      p.Name,
			Index:     p.TurnIndex,
			CardCount: p.CardCount,
			Status:    mapPlayerStatus(p.RegionCount),
		}
	}

	return result
}

func mapPlayerStatus(regionCount int64) apisnapshot.PlayerStatus {
	if regionCount == 0 {
		return apisnapshot.PlayerDead
	}

	return apisnapshot.PlayerAlive
}

// ---------------------------------------------------------------------------
// Private snapshot mapping
// ---------------------------------------------------------------------------

func (r *reader) mapPrivateSnapshot(
	ctx gamectx.GameContext,
	snap *PrivateSnapshot,
) (*apisnapshot.PlayerPrivate, error) {
	cards, err := mapCards(snap.Cards)
	if err != nil {
		return nil, fmt.Errorf("mapping cards: %w", err)
	}

	playerMission, err := r.resolveMission(ctx, snap.MissionType, snap.MissionID)
	if err != nil {
		return nil, fmt.Errorf("mapping mission: %w", err)
	}

	return &apisnapshot.PlayerPrivate{
		Cards:   cards,
		Mission: playerMission,
	}, nil
}

func mapCards(cards []sqlc.GetCardsForPlayerRow) ([]apisnapshot.CardState, error) {
	result := make([]apisnapshot.CardState, len(cards))
	for i, c := range cards {
		cardType, err := mapCardType(c.CardType)
		if err != nil {
			return nil, err
		}

		region := ""
		if c.Region.Valid {
			region = c.Region.String
		}

		result[i] = apisnapshot.CardState{
			ID:     c.ID,
			Type:   cardType,
			Region: region,
		}
	}

	return result, nil
}

func mapCardType(t sqlc.GameCardType) (apisnapshot.CardType, error) {
	switch t {
	case sqlc.GameCardTypeINFANTRY:
		return apisnapshot.CardInfantry, nil
	case sqlc.GameCardTypeCAVALRY:
		return apisnapshot.CardCavalry, nil
	case sqlc.GameCardTypeARTILLERY:
		return apisnapshot.CardArtillery, nil
	case sqlc.GameCardTypeJOLLY:
		return apisnapshot.CardJolly, nil
	default:
		return "", fmt.Errorf("unknown card type: %s", t)
	}
}

func (r *reader) resolveMission(
	ctx gamectx.GameContext,
	missionType sqlc.GameMissionType,
	missionID int64,
) (apisnapshot.PlayerMission, error) {
	switch missionType {
	case sqlc.GameMissionTypeTWOCONTINENTS:
		m, err := r.missionQuerier.GetTwoContinentsMission(ctx, missionID)
		if err != nil {
			return apisnapshot.PlayerMission{}, fmt.Errorf(
				"fetching two continents mission: %w",
				err,
			)
		}

		return apisnapshot.PlayerMission{
			Type: apisnapshot.MissionTwoContinents,
			Detail: apisnapshot.TwoContinentsMission{
				Continent1: m.Continent1,
				Continent2: m.Continent2,
			},
		}, nil
	case sqlc.GameMissionTypeTWOCONTINENTSPLUSONE:
		m, err := r.missionQuerier.GetTwoContinentsPlusOneMission(ctx, missionID)
		if err != nil {
			return apisnapshot.PlayerMission{}, fmt.Errorf(
				"fetching two continents plus one mission: %w",
				err,
			)
		}

		return apisnapshot.PlayerMission{
			Type: apisnapshot.MissionTwoContinentsPlusOne,
			Detail: apisnapshot.TwoContinentsPlusOneMission{
				Continent1: m.Continent1,
				Continent2: m.Continent2,
			},
		}, nil
	case sqlc.GameMissionTypeELIMINATEPLAYER:
		targetUser, err := r.missionQuerier.GetEliminatePlayerMission(ctx, missionID)
		if err != nil {
			return apisnapshot.PlayerMission{}, fmt.Errorf(
				"fetching eliminate player mission: %w",
				err,
			)
		}

		return apisnapshot.PlayerMission{
			Type: apisnapshot.MissionEliminatePlayer,
			Detail: apisnapshot.EliminatePlayerMission{
				TargetUserID: targetUser,
			},
		}, nil
	case sqlc.GameMissionTypeEIGHTEENTERRITORIESTWOTROOPS:
		return apisnapshot.PlayerMission{
			Type:   apisnapshot.MissionEighteenTerritoriesTwoTroops,
			Detail: apisnapshot.EighteenTerritoriesTwoTroopsMission{},
		}, nil
	case sqlc.GameMissionTypeTWENTYFOURTERRITORIES:
		return apisnapshot.PlayerMission{
			Type:   apisnapshot.MissionTwentyFourTerritories,
			Detail: apisnapshot.TwentyFourTerritoriesMission{},
		}, nil
	default:
		return apisnapshot.PlayerMission{}, fmt.Errorf("unknown mission type: %s", missionType)
	}
}
