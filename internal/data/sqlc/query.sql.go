// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createMoveLog = `-- name: CreateMoveLog :one
INSERT INTO move_log (game_id,
                      player_id,
                      phase,
                      move_data,
                      result)
VALUES ($1,
        (SELECT id FROM player WHERE game_id = $1 AND user_id = $2),
        (SELECT p.type
         FROM phase p
                  join game g on g.current_phase_id = p.id
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

func (q *Queries) CreateMoveLog(ctx context.Context, arg CreateMoveLogParams) (MoveLog, error) {
	row := q.db.QueryRow(ctx, createMoveLog,
		arg.GameID,
		arg.UserID,
		arg.MoveData,
		arg.Result,
	)
	var i MoveLog
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
UPDATE deploy_phase
SET deployable_troops = deploy_phase.deployable_troops - $2
WHERE id = (select dp.id
            from game g
                     join phase p on g.current_phase_id = p.id
                     join deploy_phase dp on p.id = dp.phase_id
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

const drawCard = `-- name: DrawCard :exec
update card
set owner_id = $2
where id = $1
`

type DrawCardParams struct {
	ID      int64
	OwnerID pgtype.Int8
}

func (q *Queries) DrawCard(ctx context.Context, arg DrawCardParams) error {
	_, err := q.db.Exec(ctx, drawCard, arg.ID, arg.OwnerID)
	return err
}

const getAvailableCards = `-- name: GetAvailableCards :many
select c.id, c.game_id, c.region_id, c.owner_id, c.card_type
from game g
         join card c on c.game_id = g.id
where g.id = $1
  and c.owner_id is null
`

func (q *Queries) GetAvailableCards(ctx context.Context, id int64) ([]Card, error) {
	rows, err := q.db.Query(ctx, getAvailableCards, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Card
	for rows.Next() {
		var i Card
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
FROM game g
         JOIN player p on g.id = p.game_id
         JOIN card c ON c.owner_id = p.id
         LEFT JOIN region r ON c.region_id = r.id
WHERE g.id = $1
  AND p.user_id = $2
`

type GetCardsForPlayerParams struct {
	ID     int64
	UserID string
}

type GetCardsForPlayerRow struct {
	ID       int64
	CardType CardType
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
from game g
         join phase p on g.current_phase_id = p.id
         join conquer_phase cp on p.id = cp.phase_id
         join region source_region on cp.source_region_id = source_region.id
         join region target_region on cp.target_region_id = target_region.id
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
FROM phase p
         JOIN GAME g on g.current_phase_id = p.id
WHERE g.id = $1
`

func (q *Queries) GetCurrentPhase(ctx context.Context, id int64) (PhaseType, error) {
	row := q.db.QueryRow(ctx, getCurrentPhase, id)
	var type_ PhaseType
	err := row.Scan(&type_)
	return type_, err
}

const getDeployableTroops = `-- name: GetDeployableTroops :one
SELECT deploy_phase.deployable_troops
FROM game
         JOIN phase ON game.current_phase_id = phase.id
         JOIN deploy_phase ON phase.id = deploy_phase.phase_id
WHERE game.id = $1
`

func (q *Queries) GetDeployableTroops(ctx context.Context, id int64) (int64, error) {
	row := q.db.QueryRow(ctx, getDeployableTroops, id)
	var deployable_troops int64
	err := row.Scan(&deployable_troops)
	return deployable_troops, err
}

const getGame = `-- name: GetGame :one
SELECT game.id, phase.type AS current_phase, phase.turn
FROM game
         JOIN phase ON game.current_phase_id = phase.id
WHERE game.id = $1
`

type GetGameRow struct {
	ID           int64
	CurrentPhase PhaseType
	Turn         int64
}

func (q *Queries) GetGame(ctx context.Context, id int64) (GetGameRow, error) {
	row := q.db.QueryRow(ctx, getGame, id)
	var i GetGameRow
	err := row.Scan(&i.ID, &i.CurrentPhase, &i.Turn)
	return i, err
}

const getMoveLogs = `-- name: GetMoveLogs :many
SELECT move_log.phase, move_log.move_data, move_log.result, move_log.created, player.user_id
FROM move_log
         JOIN player ON player.id = player_id
WHERE move_log.game_id = $1
ORDER BY created DESC
LIMIT $2::bigint
`

type GetMoveLogsParams struct {
	GameID  int64
	MaxLogs int64
}

type GetMoveLogsRow struct {
	Phase    PhaseType
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
FROM player
WHERE player.game_id = $1
  AND player.turn_index = ((1 + (SELECT p.turn
                                 FROM game g
                                          JOIN phase p on g.current_phase_id = p.id
                                 WHERE g.id = $1))
    % (SELECT COUNT (player.id) FROM player WHERE player.game_id = $1))
`

func (q *Queries) GetNextPlayer(ctx context.Context, gameID int64) (Player, error) {
	row := q.db.QueryRow(ctx, getNextPlayer, gameID)
	var i Player
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
FROM player
WHERE user_id = $1
`

func (q *Queries) GetPlayerByUserId(ctx context.Context, userID string) (Player, error) {
	row := q.db.QueryRow(ctx, getPlayerByUserId, userID)
	var i Player
	err := row.Scan(
		&i.ID,
		&i.GameID,
		&i.Name,
		&i.UserID,
		&i.TurnIndex,
	)
	return i, err
}

const getPlayerRegionsFromRegion = `-- name: GetPlayerRegionsFromRegion :one
SELECT p.user_id, COUNT(r.id) as region_count
FROM player p
         JOIN region r on r.player_id = p.id
         JOIN region this_region on this_region.player_id = p.id
WHERE p.game_id = $1
  AND this_region.external_reference = $2
GROUP BY p.user_id
`

type GetPlayerRegionsFromRegionParams struct {
	GameID            int64
	ExternalReference string
}

type GetPlayerRegionsFromRegionRow struct {
	UserID      string
	RegionCount int64
}

func (q *Queries) GetPlayerRegionsFromRegion(ctx context.Context, arg GetPlayerRegionsFromRegionParams) (GetPlayerRegionsFromRegionRow, error) {
	row := q.db.QueryRow(ctx, getPlayerRegionsFromRegion, arg.GameID, arg.ExternalReference)
	var i GetPlayerRegionsFromRegionRow
	err := row.Scan(&i.UserID, &i.RegionCount)
	return i, err
}

const getPlayersByGame = `-- name: GetPlayersByGame :many
SELECT id, game_id, name, user_id, turn_index
FROM player
WHERE game_id = $1
`

func (q *Queries) GetPlayersByGame(ctx context.Context, gameID int64) ([]Player, error) {
	rows, err := q.db.Query(ctx, getPlayersByGame, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Player
	for rows.Next() {
		var i Player
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
FROM player p
         LEFT JOIN card c on p.id = c.owner_id
         LEFT JOIN region r on r.player_id = p.id
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
FROM region r
         JOIN player p on r.player_id = p.id
         JOIN game g on p.game_id = g.id
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

const grantRegionTroops = `-- name: GrantRegionTroops :exec
UPDATE region
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
            from game g
                     join phase p on p.game_id = g.id
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
UPDATE region
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
	CardType CardType
}

const insertConquerPhase = `-- name: InsertConquerPhase :one
INSERT INTO conquer_phase(phase_id, source_region_id, target_region_id, minimum_troops)
VALUES ($1,
        (select r.id
         from game g
                  join player p on g.id = p.game_id
                  join region r on p.id = r.player_id
         where g.id = $2
           and r.external_reference = $3),
        (select r.id
         from game g
                  join player p on g.id = p.game_id
                  join region r on p.id = r.player_id
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

func (q *Queries) InsertConquerPhase(ctx context.Context, arg InsertConquerPhaseParams) (ConquerPhase, error) {
	row := q.db.QueryRow(ctx, insertConquerPhase,
		arg.PhaseID,
		arg.ID,
		arg.ExternalReference,
		arg.ExternalReference_2,
		arg.MinimumTroops,
	)
	var i ConquerPhase
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
INSERT INTO deploy_phase (phase_id, deployable_troops)
VALUES ($1, $2) RETURNING id, phase_id, deployable_troops
`

type InsertDeployPhaseParams struct {
	PhaseID          int64
	DeployableTroops int64
}

func (q *Queries) InsertDeployPhase(ctx context.Context, arg InsertDeployPhaseParams) (DeployPhase, error) {
	row := q.db.QueryRow(ctx, insertDeployPhase, arg.PhaseID, arg.DeployableTroops)
	var i DeployPhase
	err := row.Scan(&i.ID, &i.PhaseID, &i.DeployableTroops)
	return i, err
}

const insertGame = `-- name: InsertGame :one
INSERT INTO game DEFAULT
VALUES
RETURNING id, current_phase_id, winner_player_id
`

func (q *Queries) InsertGame(ctx context.Context) (Game, error) {
	row := q.db.QueryRow(ctx, insertGame)
	var i Game
	err := row.Scan(&i.ID, &i.CurrentPhaseID, &i.WinnerPlayerID)
	return i, err
}

const insertPhase = `-- name: InsertPhase :one
INSERT INTO phase (game_id, type, turn)
VALUES ($1, $2, $3) RETURNING id, game_id, type, turn
`

type InsertPhaseParams struct {
	GameID int64
	Type   PhaseType
	Turn   int64
}

func (q *Queries) InsertPhase(ctx context.Context, arg InsertPhaseParams) (Phase, error) {
	row := q.db.QueryRow(ctx, insertPhase, arg.GameID, arg.Type, arg.Turn)
	var i Phase
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

const setGamePhase = `-- name: SetGamePhase :exec
UPDATE game
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
UPDATE card
SET owner_id = (SELECT id from player WHERE player.user_id = $2::text AND player.game_id = $1)
WHERE owner_id = (SELECT id from player WHERE player.user_id = $3::text AND player.game_id = $1)
`

type TransferCardsOwnershipParams struct {
	GameID int64
	To     string
	From   string
}

func (q *Queries) TransferCardsOwnership(ctx context.Context, arg TransferCardsOwnershipParams) error {
	_, err := q.db.Exec(ctx, transferCardsOwnership, arg.GameID, arg.To, arg.From)
	return err
}

const unlinkCardsFromOwner = `-- name: UnlinkCardsFromOwner :exec
UPDATE card
SET owner_id = NULL
WHERE id = ANY ($1::bigint[])
`

func (q *Queries) UnlinkCardsFromOwner(ctx context.Context, cards []int64) error {
	_, err := q.db.Exec(ctx, unlinkCardsFromOwner, cards)
	return err
}

const updateRegionOwner = `-- name: UpdateRegionOwner :exec
UPDATE region
SET player_id = (SELECT player.id FROM player WHERE user_id = $1 AND game_id = $2)
WHERE region.id = $3
`

type UpdateRegionOwnerParams struct {
	UserID string
	GameID int64
	ID     int64
}

func (q *Queries) UpdateRegionOwner(ctx context.Context, arg UpdateRegionOwnerParams) error {
	_, err := q.db.Exec(ctx, updateRegionOwner, arg.UserID, arg.GameID, arg.ID)
	return err
}
