// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: query.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

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
VALUES ($1, $2)
RETURNING id, phase_id, deployable_troops
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
RETURNING id, current_phase_id
`

func (q *Queries) InsertGame(ctx context.Context) (Game, error) {
	row := q.db.QueryRow(ctx, insertGame)
	var i Game
	err := row.Scan(&i.ID, &i.CurrentPhaseID)
	return i, err
}

const insertPhase = `-- name: InsertPhase :one
INSERT INTO phase (game_id, type, turn)
VALUES ($1, $2, $3)
RETURNING id, game_id, type, turn
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
