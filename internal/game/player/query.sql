-- name: GetPlayersByGameId :many
SELECT * FROM player WHERE game_id = $1;

-- name: InsertPlayers :copyfrom
INSERT INTO player (game_id, user_id) VALUES ($1, $2);
