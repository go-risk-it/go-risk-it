-- name: InsertRegions :copyfrom
INSERT INTO region (external_reference, player_id, troops)
VALUES ($1, $2, $3);

-- name: GetRegionsByGame :many
SELECT r.id, r.external_reference, r.troops, p.user_id
FROM region r
         JOIN player p on r.player_id = p.id
         JOIN game g on p.game_id = g.id
WHERE g.id = $1;

-- name: IncreaseRegionTroops :exec
UPDATE region
SET troops = troops + $2
WHERE id = $1;

-- name: UpdateRegionOwner :exec
UPDATE region
SET player_id = (SELECT player.id FROM player WHERE user_id = $1 AND game_id = $2)
WHERE region.id = $3;