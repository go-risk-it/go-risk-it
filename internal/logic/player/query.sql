-- name: GetPlayersByGame :many
SELECT *
FROM player
WHERE game_id = $1;

-- name: InsertPlayers :copyfrom
INSERT INTO player (game_id, user_id, turn_index, troops_to_deploy)
VALUES ($1, $2, $3, $4);
