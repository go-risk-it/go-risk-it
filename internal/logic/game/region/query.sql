-- name: InsertRegions :copyfrom
INSERT INTO game.region (external_reference, player_id, troops)
VALUES ($1, $2, $3);

-- name: GetRegionsByGame :many
SELECT r.id, r.external_reference, r.troops, p.user_id
FROM game.region r
         JOIN game.player p on r.player_id = p.id
         JOIN game.game g on p.game_id = g.id
WHERE g.id = $1;

-- name: GetRegionsByPlayer :many
SELECT r.*
FROM game.region r
         JOIN game.player p on r.player_id = p.id
WHERE p.id = $1;

-- name: IncreaseRegionTroops :exec
UPDATE game.region
SET troops = troops + $2
WHERE id = $1;

-- name: UpdateRegionOwner :one
WITH old_value AS (
    SELECT player_id FROM game.region WHERE id = sqlc.arg(conquered_region_id)
)
UPDATE game.region
SET player_id = (SELECT player.id FROM game.player WHERE user_id = sqlc.arg(new_owner_user_id) AND game_id = $1)
WHERE region.id = sqlc.arg(conquered_region_id)
RETURNING (SELECT player_id FROM old_value) AS old_player_id;