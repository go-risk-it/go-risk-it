-- name: InsertRegions :copyfrom
INSERT INTO region (external_reference, player_id, troops)
VALUES ($1, $2, $3);

-- name: GetRegionsByGame :many
SELECT r.external_reference, r.troops, p.user_id as player_name
FROM region r
         JOIN player p on r.player_id = p.id
         JOIN game g on p.game_id = g.id
WHERE g.id = $1;
