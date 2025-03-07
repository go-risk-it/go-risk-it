// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: query.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const assignGameWinner = `-- name: AssignGameWinner :exec
UPDATE game.game
SET winner_player_id = $1
WHERE id = $2
`

type AssignGameWinnerParams struct {
	WinnerPlayerID pgtype.Int8
	GameID         int64
}

func (q *Queries) AssignGameWinner(ctx context.Context, arg AssignGameWinnerParams) error {
	_, err := q.db.Exec(ctx, assignGameWinner, arg.WinnerPlayerID, arg.GameID)
	return err
}

const createMoveLog = `-- name: CreateMoveLog :one
INSERT INTO game.move_log (game_id,
                      player_id,
                      phase,
                      move_data,
                      result)
VALUES ($1,
        (SELECT id FROM game.player WHERE game_id = $1 AND user_id = $2),
        (SELECT p.type
         FROM game.phase p
                  join game.game g on g.current_phase_id = p.id
         WHERE g.id = $1),
        $3,
        $4)
RETURNING id, game_id, player_id, phase, move_data, result, created
`

type CreateMoveLogParams struct {
	GameID   int64
	UserID   string
	MoveData []byte
	Result   []byte
}

func (q *Queries) CreateMoveLog(ctx context.Context, arg CreateMoveLogParams) (GameMoveLog, error) {
	row := q.db.QueryRow(ctx, createMoveLog,
		arg.GameID,
		arg.UserID,
		arg.MoveData,
		arg.Result,
	)
	var i GameMoveLog
	err := row.Scan(
		&i.ID,
		&i.GameID,
		&i.PlayerID,
		&i.Phase,
		&i.MoveData,
		&i.Result,
		&i.Created,
	)
	return i, err
}

const decreaseDeployableTroops = `-- name: DecreaseDeployableTroops :exec
UPDATE game.deploy_phase
SET deployable_troops = game.deploy_phase.deployable_troops - $2
WHERE id = (select dp.id
            from game.game g
                     join game.phase p on g.current_phase_id = p.id
                     join game.deploy_phase dp on p.id = dp.phase_id
            where g.id = $1)
`

type DecreaseDeployableTroopsParams struct {
	ID               int64
	DeployableTroops int64
}

func (q *Queries) DecreaseDeployableTroops(ctx context.Context, arg DecreaseDeployableTroopsParams) error {
	_, err := q.db.Exec(ctx, decreaseDeployableTroops, arg.ID, arg.DeployableTroops)
	return err
}

const deleteSpuriousEliminatePlayerMissions = `-- name: DeleteSpuriousEliminatePlayerMissions :exec
DELETE
FROM game.eliminate_player_mission
WHERE mission_id in (SELECT m.id
                     FROM game.mission m
                              JOIN game.player p on m.player_id = p.id
                     WHERE p.game_id = $1
                       AND m.type = 'TWENTY_FOUR_TERRITORIES')
`

func (q *Queries) DeleteSpuriousEliminatePlayerMissions(ctx context.Context, gameID int64) error {
	_, err := q.db.Exec(ctx, deleteSpuriousEliminatePlayerMissions, gameID)
	return err
}

const drawCard = `-- name: DrawCard :exec
update game.card
set owner_id = (select game.player.id from game.player where game.player.user_id = $2 and game.player.game_id = $3)
where game.card.id = $1
`

type DrawCardParams struct {
	ID     int64
	UserID string
	GameID int64
}

func (q *Queries) DrawCard(ctx context.Context, arg DrawCardParams) error {
	_, err := q.db.Exec(ctx, drawCard, arg.ID, arg.UserID, arg.GameID)
	return err
}

const getAvailableCards = `-- name: GetAvailableCards :many
select c.id, c.game_id, c.region_id, c.owner_id, c.card_type
from game.game g
         join game.card c on c.game_id = g.id
where g.id = $1
  and c.owner_id is null
`

func (q *Queries) GetAvailableCards(ctx context.Context, id int64) ([]GameCard, error) {
	rows, err := q.db.Query(ctx, getAvailableCards, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GameCard
	for rows.Next() {
		var i GameCard
		if err := rows.Scan(
			&i.ID,
			&i.GameID,
			&i.RegionID,
			&i.OwnerID,
			&i.CardType,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getCardsForPlayer = `-- name: GetCardsForPlayer :many
SELECT c.id, c.card_type, r.external_reference as region
FROM game.game g
         JOIN game.player p on g.id = p.game_id
         JOIN game.card c ON c.owner_id = p.id
         LEFT JOIN game.region r ON c.region_id = r.id
WHERE g.id = $1
  AND p.user_id = $2
`

type GetCardsForPlayerParams struct {
	ID     int64
	UserID string
}

type GetCardsForPlayerRow struct {
	ID       int64
	CardType GameCardType
	Region   pgtype.Text
}

func (q *Queries) GetCardsForPlayer(ctx context.Context, arg GetCardsForPlayerParams) ([]GetCardsForPlayerRow, error) {
	rows, err := q.db.Query(ctx, getCardsForPlayer, arg.ID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetCardsForPlayerRow
	for rows.Next() {
		var i GetCardsForPlayerRow
		if err := rows.Scan(&i.ID, &i.CardType, &i.Region); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getConquerPhaseState = `-- name: GetConquerPhaseState :one
select source_region.external_reference as source_region,
       target_region.external_reference as target_region,
       cp.minimum_troops
from game.game g
         join game.phase p on g.current_phase_id = p.id
         join game.conquer_phase cp on p.id = cp.phase_id
         join game.region source_region on cp.source_region_id = source_region.id
         join game.region target_region on cp.target_region_id = target_region.id
where g.id = $1
`

type GetConquerPhaseStateRow struct {
	SourceRegion  string
	TargetRegion  string
	MinimumTroops int64
}

func (q *Queries) GetConquerPhaseState(ctx context.Context, id int64) (GetConquerPhaseStateRow, error) {
	row := q.db.QueryRow(ctx, getConquerPhaseState, id)
	var i GetConquerPhaseStateRow
	err := row.Scan(&i.SourceRegion, &i.TargetRegion, &i.MinimumTroops)
	return i, err
}

const getCurrentPhase = `-- name: GetCurrentPhase :one
SELECT p.type
FROM game.phase p
         JOIN game.game g on g.current_phase_id = p.id
WHERE g.id = $1
`

func (q *Queries) GetCurrentPhase(ctx context.Context, id int64) (GamePhaseType, error) {
	row := q.db.QueryRow(ctx, getCurrentPhase, id)
	var type_ GamePhaseType
	err := row.Scan(&type_)
	return type_, err
}

const getCurrentPlayer = `-- name: GetCurrentPlayer :one
SELECT id, game_id, name, user_id, turn_index
FROM game.player
WHERE game.player.game_id = $1
  AND game.player.turn_index = ((SELECT p.turn
                            FROM game.game g
                                     JOIN game.phase p on g.current_phase_id = p.id
                            WHERE g.id = $1)
    % (SELECT COUNT (player.id) FROM game.player WHERE player.game_id = $1))
`

func (q *Queries) GetCurrentPlayer(ctx context.Context, gameID int64) (GamePlayer, error) {
	row := q.db.QueryRow(ctx, getCurrentPlayer, gameID)
	var i GamePlayer
	err := row.Scan(
		&i.ID,
		&i.GameID,
		&i.Name,
		&i.UserID,
		&i.TurnIndex,
	)
	return i, err
}

const getDeployableTroops = `-- name: GetDeployableTroops :one
SELECT deploy_phase.deployable_troops
FROM game.game
         JOIN game.phase ON game.current_phase_id = game.phase.id
         JOIN game.deploy_phase ON game.phase.id = game.deploy_phase.phase_id
WHERE game.id = $1
`

func (q *Queries) GetDeployableTroops(ctx context.Context, id int64) (int64, error) {
	row := q.db.QueryRow(ctx, getDeployableTroops, id)
	var deployable_troops int64
	err := row.Scan(&deployable_troops)
	return deployable_troops, err
}

const getEliminatePlayerMission = `-- name: GetEliminatePlayerMission :one
SELECT mission_id, target_player_id
FROM game.eliminate_player_mission
WHERE mission_id = $1
`

func (q *Queries) GetEliminatePlayerMission(ctx context.Context, missionID int64) (GameEliminatePlayerMission, error) {
	row := q.db.QueryRow(ctx, getEliminatePlayerMission, missionID)
	var i GameEliminatePlayerMission
	err := row.Scan(&i.MissionID, &i.TargetPlayerID)
	return i, err
}

const getGame = `-- name: GetGame :one
SELECT g.id, p.type AS current_phase, p.turn, winner_player.user_id AS winner_user_id
FROM game.game g
         JOIN game.phase p ON g.current_phase_id = p.id
         LEFT JOIN game.player winner_player ON g.winner_player_id = winner_player.id
WHERE g.id = $1
`

type GetGameRow struct {
	ID           int64
	CurrentPhase GamePhaseType
	Turn         int64
	WinnerUserID pgtype.Text
}

func (q *Queries) GetGame(ctx context.Context, id int64) (GetGameRow, error) {
	row := q.db.QueryRow(ctx, getGame, id)
	var i GetGameRow
	err := row.Scan(
		&i.ID,
		&i.CurrentPhase,
		&i.Turn,
		&i.WinnerUserID,
	)
	return i, err
}

const getMission = `-- name: GetMission :one
SELECT m.id, m.player_id, m.type
FROM game.mission m
         JOIN game.player p ON m.player_id = p.id
WHERE p.game_id = $1
  AND p.user_id = $2
`

type GetMissionParams struct {
	GameID int64
	UserID string
}

func (q *Queries) GetMission(ctx context.Context, arg GetMissionParams) (GameMission, error) {
	row := q.db.QueryRow(ctx, getMission, arg.GameID, arg.UserID)
	var i GameMission
	err := row.Scan(&i.ID, &i.PlayerID, &i.Type)
	return i, err
}

const getMoveLogs = `-- name: GetMoveLogs :many
SELECT move_log.phase, move_log.move_data, move_log.result, move_log.created, player.user_id
FROM game.move_log
         JOIN game.player ON game.player.id = player_id
WHERE game.move_log.game_id = $1
ORDER BY created DESC
LIMIT $2::bigint
`

type GetMoveLogsParams struct {
	GameID  int64
	MaxLogs int64
}

type GetMoveLogsRow struct {
	Phase    GamePhaseType
	MoveData []byte
	Result   []byte
	Created  pgtype.Timestamptz
	UserID   string
}

func (q *Queries) GetMoveLogs(ctx context.Context, arg GetMoveLogsParams) ([]GetMoveLogsRow, error) {
	rows, err := q.db.Query(ctx, getMoveLogs, arg.GameID, arg.MaxLogs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetMoveLogsRow
	for rows.Next() {
		var i GetMoveLogsRow
		if err := rows.Scan(
			&i.Phase,
			&i.MoveData,
			&i.Result,
			&i.Created,
			&i.UserID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getNextPlayer = `-- name: GetNextPlayer :one
SELECT id, game_id, name, user_id, turn_index
FROM game.player
WHERE game.player.game_id = $1
  AND game.player.turn_index = ((1 + (SELECT p.turn
                                 FROM game.game g
                                          JOIN game.phase p on g.current_phase_id = p.id
                                 WHERE g.id = $1))
    % (SELECT COUNT (game.player.id) FROM game.player WHERE game.player.game_id = $1))
`

func (q *Queries) GetNextPlayer(ctx context.Context, gameID int64) (GamePlayer, error) {
	row := q.db.QueryRow(ctx, getNextPlayer, gameID)
	var i GamePlayer
	err := row.Scan(
		&i.ID,
		&i.GameID,
		&i.Name,
		&i.UserID,
		&i.TurnIndex,
	)
	return i, err
}

const getPlayerByUserId = `-- name: GetPlayerByUserId :one
SELECT id, game_id, name, user_id, turn_index
FROM game.player
WHERE user_id = $1
`

func (q *Queries) GetPlayerByUserId(ctx context.Context, userID string) (GamePlayer, error) {
	row := q.db.QueryRow(ctx, getPlayerByUserId, userID)
	var i GamePlayer
	err := row.Scan(
		&i.ID,
		&i.GameID,
		&i.Name,
		&i.UserID,
		&i.TurnIndex,
	)
	return i, err
}

const getPlayersByGame = `-- name: GetPlayersByGame :many
SELECT id, game_id, name, user_id, turn_index
FROM game.player
WHERE game_id = $1
`

func (q *Queries) GetPlayersByGame(ctx context.Context, gameID int64) ([]GamePlayer, error) {
	rows, err := q.db.Query(ctx, getPlayersByGame, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GamePlayer
	for rows.Next() {
		var i GamePlayer
		if err := rows.Scan(
			&i.ID,
			&i.GameID,
			&i.Name,
			&i.UserID,
			&i.TurnIndex,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPlayersState = `-- name: GetPlayersState :many
SELECT p.user_id, p.name, p.turn_index, COUNT(distinct c.id) as card_count, COUNT(distinct r.id) as region_count
FROM game.player p
         LEFT JOIN game.card c on p.id = c.owner_id
         LEFT JOIN game.region r on r.player_id = p.id
WHERE p.game_id = $1
GROUP BY p.id
ORDER BY p.turn_index
`

type GetPlayersStateRow struct {
	UserID      string
	Name        string
	TurnIndex   int64
	CardCount   int64
	RegionCount int64
}

func (q *Queries) GetPlayersState(ctx context.Context, gameID int64) ([]GetPlayersStateRow, error) {
	rows, err := q.db.Query(ctx, getPlayersState, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPlayersStateRow
	for rows.Next() {
		var i GetPlayersStateRow
		if err := rows.Scan(
			&i.UserID,
			&i.Name,
			&i.TurnIndex,
			&i.CardCount,
			&i.RegionCount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRegionsByGame = `-- name: GetRegionsByGame :many
SELECT r.id, r.external_reference, r.troops, p.user_id
FROM game.region r
         JOIN game.player p on r.player_id = p.id
         JOIN game.game g on p.game_id = g.id
WHERE g.id = $1
`

type GetRegionsByGameRow struct {
	ID                int64
	ExternalReference string
	Troops            int64
	UserID            string
}

func (q *Queries) GetRegionsByGame(ctx context.Context, id int64) ([]GetRegionsByGameRow, error) {
	rows, err := q.db.Query(ctx, getRegionsByGame, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetRegionsByGameRow
	for rows.Next() {
		var i GetRegionsByGameRow
		if err := rows.Scan(
			&i.ID,
			&i.ExternalReference,
			&i.Troops,
			&i.UserID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRegionsByPlayer = `-- name: GetRegionsByPlayer :many
SELECT r.id, r.external_reference, r.player_id, r.troops
FROM game.region r
         JOIN game.player p on r.player_id = p.id
WHERE p.id = $1
`

func (q *Queries) GetRegionsByPlayer(ctx context.Context, id int64) ([]GameRegion, error) {
	rows, err := q.db.Query(ctx, getRegionsByPlayer, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GameRegion
	for rows.Next() {
		var i GameRegion
		if err := rows.Scan(
			&i.ID,
			&i.ExternalReference,
			&i.PlayerID,
			&i.Troops,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTwoContinentsMission = `-- name: GetTwoContinentsMission :one
SELECT mission_id, continent_1, continent_2
FROM game.two_continents_mission
WHERE mission_id = $1
`

func (q *Queries) GetTwoContinentsMission(ctx context.Context, missionID int64) (GameTwoContinentsMission, error) {
	row := q.db.QueryRow(ctx, getTwoContinentsMission, missionID)
	var i GameTwoContinentsMission
	err := row.Scan(&i.MissionID, &i.Continent1, &i.Continent2)
	return i, err
}

const getTwoContinentsPlusOneMission = `-- name: GetTwoContinentsPlusOneMission :one
SELECT mission_id, continent_1, continent_2
FROM game.two_continents_plus_one_mission
WHERE mission_id = $1
`

func (q *Queries) GetTwoContinentsPlusOneMission(ctx context.Context, missionID int64) (GameTwoContinentsPlusOneMission, error) {
	row := q.db.QueryRow(ctx, getTwoContinentsPlusOneMission, missionID)
	var i GameTwoContinentsPlusOneMission
	err := row.Scan(&i.MissionID, &i.Continent1, &i.Continent2)
	return i, err
}

const getUserGames = `-- name: GetUserGames :many
SELECT DISTINCT g.id
FROM game.game g
         JOIN game.player p on g.id = p.game_id
WHERE p.user_id = $1
  and g.winner_player_id IS NULL
`

func (q *Queries) GetUserGames(ctx context.Context, userID string) ([]int64, error) {
	rows, err := q.db.Query(ctx, getUserGames, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const grantRegionTroops = `-- name: GrantRegionTroops :exec
UPDATE game.region
set troops = troops + $1
where id = ANY ($2::bigint[])
`

type GrantRegionTroopsParams struct {
	Troops  int64
	Regions []int64
}

func (q *Queries) GrantRegionTroops(ctx context.Context, arg GrantRegionTroopsParams) error {
	_, err := q.db.Exec(ctx, grantRegionTroops, arg.Troops, arg.Regions)
	return err
}

const hasConqueredInTurn = `-- name: HasConqueredInTurn :one
select exists
           (select p.id
            from game.game g
                     join game.phase p on p.game_id = g.id
            where g.id = $1
              and p.type = 'CONQUER'
              and p.turn = $2)
`

type HasConqueredInTurnParams struct {
	ID   int64
	Turn int64
}

func (q *Queries) HasConqueredInTurn(ctx context.Context, arg HasConqueredInTurnParams) (bool, error) {
	row := q.db.QueryRow(ctx, hasConqueredInTurn, arg.ID, arg.Turn)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const increaseRegionTroops = `-- name: IncreaseRegionTroops :exec
UPDATE game.region
SET troops = troops + $2
WHERE id = $1
`

type IncreaseRegionTroopsParams struct {
	ID     int64
	Troops int64
}

func (q *Queries) IncreaseRegionTroops(ctx context.Context, arg IncreaseRegionTroopsParams) error {
	_, err := q.db.Exec(ctx, increaseRegionTroops, arg.ID, arg.Troops)
	return err
}

type InsertCardsParams struct {
	RegionID pgtype.Int8
	GameID   int64
	CardType GameCardType
}

const insertConquerPhase = `-- name: InsertConquerPhase :one
INSERT INTO game.conquer_phase(phase_id, source_region_id, target_region_id, minimum_troops)
VALUES ($1,
        (select r.id
         from game.game g
                  join game.player p on g.id = p.game_id
                  join game.region r on p.id = r.player_id
         where g.id = $2
           and r.external_reference = $3),
        (select r.id
         from game.game g
                  join game.player p on g.id = p.game_id
                  join game.region r on p.id = r.player_id
         where g.id = $2
           and r.external_reference = $4),
        $5)
RETURNING id, phase_id, source_region_id, target_region_id, minimum_troops
`

type InsertConquerPhaseParams struct {
	PhaseID             int64
	ID                  int64
	ExternalReference   string
	ExternalReference_2 string
	MinimumTroops       int64
}

func (q *Queries) InsertConquerPhase(ctx context.Context, arg InsertConquerPhaseParams) (GameConquerPhase, error) {
	row := q.db.QueryRow(ctx, insertConquerPhase,
		arg.PhaseID,
		arg.ID,
		arg.ExternalReference,
		arg.ExternalReference_2,
		arg.MinimumTroops,
	)
	var i GameConquerPhase
	err := row.Scan(
		&i.ID,
		&i.PhaseID,
		&i.SourceRegionID,
		&i.TargetRegionID,
		&i.MinimumTroops,
	)
	return i, err
}

const insertDeployPhase = `-- name: InsertDeployPhase :one
INSERT INTO game.deploy_phase (phase_id, deployable_troops)
VALUES ($1, $2) RETURNING id, phase_id, deployable_troops
`

type InsertDeployPhaseParams struct {
	PhaseID          int64
	DeployableTroops int64
}

func (q *Queries) InsertDeployPhase(ctx context.Context, arg InsertDeployPhaseParams) (GameDeployPhase, error) {
	row := q.db.QueryRow(ctx, insertDeployPhase, arg.PhaseID, arg.DeployableTroops)
	var i GameDeployPhase
	err := row.Scan(&i.ID, &i.PhaseID, &i.DeployableTroops)
	return i, err
}

const insertEliminatePlayerMission = `-- name: InsertEliminatePlayerMission :exec
INSERT INTO game.eliminate_player_mission (mission_id, target_player_id)
VALUES ($1, $2)
`

type InsertEliminatePlayerMissionParams struct {
	MissionID      int64
	TargetPlayerID int64
}

func (q *Queries) InsertEliminatePlayerMission(ctx context.Context, arg InsertEliminatePlayerMissionParams) error {
	_, err := q.db.Exec(ctx, insertEliminatePlayerMission, arg.MissionID, arg.TargetPlayerID)
	return err
}

const insertGame = `-- name: InsertGame :one
INSERT INTO game.game DEFAULT
VALUES
RETURNING id, current_phase_id, winner_player_id
`

func (q *Queries) InsertGame(ctx context.Context) (GameGame, error) {
	row := q.db.QueryRow(ctx, insertGame)
	var i GameGame
	err := row.Scan(&i.ID, &i.CurrentPhaseID, &i.WinnerPlayerID)
	return i, err
}

const insertMission = `-- name: InsertMission :one
INSERT INTO game.mission (player_id, type)
VALUES ($1, $2)
RETURNING id
`

type InsertMissionParams struct {
	PlayerID int64
	Type     GameMissionType
}

func (q *Queries) InsertMission(ctx context.Context, arg InsertMissionParams) (int64, error) {
	row := q.db.QueryRow(ctx, insertMission, arg.PlayerID, arg.Type)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const insertPhase = `-- name: InsertPhase :one
INSERT INTO game.phase (game_id, type, turn)
VALUES ($1, $2, $3) RETURNING id, game_id, type, turn
`

type InsertPhaseParams struct {
	GameID int64
	Type   GamePhaseType
	Turn   int64
}

func (q *Queries) InsertPhase(ctx context.Context, arg InsertPhaseParams) (GamePhase, error) {
	row := q.db.QueryRow(ctx, insertPhase, arg.GameID, arg.Type, arg.Turn)
	var i GamePhase
	err := row.Scan(
		&i.ID,
		&i.GameID,
		&i.Type,
		&i.Turn,
	)
	return i, err
}

type InsertPlayersParams struct {
	GameID    int64
	UserID    string
	Name      string
	TurnIndex int64
}

type InsertRegionsParams struct {
	ExternalReference string
	PlayerID          int64
	Troops            int64
}

const insertTwoContinentsMission = `-- name: InsertTwoContinentsMission :exec
INSERT INTO game.two_continents_mission (mission_id, continent_1, continent_2)
VALUES ($1, $2, $3)
`

type InsertTwoContinentsMissionParams struct {
	MissionID  int64
	Continent1 string
	Continent2 string
}

func (q *Queries) InsertTwoContinentsMission(ctx context.Context, arg InsertTwoContinentsMissionParams) error {
	_, err := q.db.Exec(ctx, insertTwoContinentsMission, arg.MissionID, arg.Continent1, arg.Continent2)
	return err
}

const insertTwoContinentsPlusOneMission = `-- name: InsertTwoContinentsPlusOneMission :exec
INSERT INTO game.two_continents_plus_one_mission (mission_id, continent_1, continent_2)
VALUES ($1, $2, $3)
`

type InsertTwoContinentsPlusOneMissionParams struct {
	MissionID  int64
	Continent1 string
	Continent2 string
}

func (q *Queries) InsertTwoContinentsPlusOneMission(ctx context.Context, arg InsertTwoContinentsPlusOneMissionParams) error {
	_, err := q.db.Exec(ctx, insertTwoContinentsPlusOneMission, arg.MissionID, arg.Continent1, arg.Continent2)
	return err
}

const reassignMissions = `-- name: ReassignMissions :exec
UPDATE game.mission
SET type = 'TWENTY_FOUR_TERRITORIES'
WHERE id in (SELECT m.id
             FROM game.mission m
                      JOIN game.player p on m.player_id = p.id
                      JOIN game.eliminate_player_mission em on m.id = em.mission_id
             WHERE p.game_id = $1
               AND em.target_player_id = $3
               AND p.user_id <> $2)
`

type ReassignMissionsParams struct {
	GameID             int64
	UserID             string
	EliminatedPlayerID int64
}

func (q *Queries) ReassignMissions(ctx context.Context, arg ReassignMissionsParams) error {
	_, err := q.db.Exec(ctx, reassignMissions, arg.GameID, arg.UserID, arg.EliminatedPlayerID)
	return err
}

const setGamePhase = `-- name: SetGamePhase :exec
UPDATE game.game
SET current_phase_id = $2
WHERE id = $1
`

type SetGamePhaseParams struct {
	ID             int64
	CurrentPhaseID pgtype.Int8
}

func (q *Queries) SetGamePhase(ctx context.Context, arg SetGamePhaseParams) error {
	_, err := q.db.Exec(ctx, setGamePhase, arg.ID, arg.CurrentPhaseID)
	return err
}

const transferCardsOwnership = `-- name: TransferCardsOwnership :exec
UPDATE game.card
SET owner_id = (SELECT id from game.player WHERE player.user_id = $2::text AND player.game_id = $1)
WHERE owner_id = $3
`

type TransferCardsOwnershipParams struct {
	GameID int64
	To     string
	From   pgtype.Int8
}

func (q *Queries) TransferCardsOwnership(ctx context.Context, arg TransferCardsOwnershipParams) error {
	_, err := q.db.Exec(ctx, transferCardsOwnership, arg.GameID, arg.To, arg.From)
	return err
}

const unlinkCardsFromOwner = `-- name: UnlinkCardsFromOwner :exec
UPDATE game.card
SET owner_id = NULL
WHERE id = ANY ($1::bigint[])
`

func (q *Queries) UnlinkCardsFromOwner(ctx context.Context, cards []int64) error {
	_, err := q.db.Exec(ctx, unlinkCardsFromOwner, cards)
	return err
}

const updateRegionOwner = `-- name: UpdateRegionOwner :one
WITH old_value AS (
    SELECT player_id FROM game.region WHERE id = $3
)
UPDATE game.region
SET player_id = (SELECT player.id FROM game.player WHERE user_id = $2 AND game_id = $1)
WHERE region.id = $3
RETURNING (SELECT player_id FROM old_value) AS old_player_id
`

type UpdateRegionOwnerParams struct {
	GameID            int64
	NewOwnerUserID    string
	ConqueredRegionID int64
}

func (q *Queries) UpdateRegionOwner(ctx context.Context, arg UpdateRegionOwnerParams) (int64, error) {
	row := q.db.QueryRow(ctx, updateRegionOwner, arg.GameID, arg.NewOwnerUserID, arg.ConqueredRegionID)
	var old_player_id int64
	err := row.Scan(&old_player_id)
	return old_player_id, err
}
