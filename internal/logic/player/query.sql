-- name: GetPlayersByGame :many
SELECT *
FROM player
WHERE game_id = $1;

-- name: GetPlayerByUserId :one
SELECT *
FROM player
WHERE user_id = $1;

-- name: InsertPlayers :copyfrom
INSERT INTO player (game_id, user_id, name, turn_index)
VALUES ($1, $2, $3, $4);

