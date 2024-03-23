-- name: GetPlayersByGame :many
SELECT *
FROM player
WHERE game_id = $1;

-- name: GetPlayerByUserId :one
SELECT *
FROM player
WHERE user_id = $1;

-- name: InsertPlayers :copyfrom
INSERT INTO player (game_id, user_id, turn_index, deployable_troops)
VALUES ($1, $2, $3, $4);

-- name: DecreaseDeployableTroops :exec
UPDATE player
SET deployable_troops = deployable_troops - $2
WHERE id = $1;

