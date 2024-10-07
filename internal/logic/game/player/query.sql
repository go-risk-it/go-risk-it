-- name: GetPlayersByGame :many
SELECT *
FROM player
WHERE game_id = $1;

-- name: GetPlayersState :many
SELECT p.user_id, p.name, p.turn_index, COUNT(c.id) as card_count
FROM player p
         LEFT JOIN card c on p.id = c.owner_id
WHERE p.game_id = $1
GROUP BY p.id;

-- name: GetPlayerByUserId :one
SELECT *
FROM player
WHERE user_id = $1;

-- name: InsertPlayers :copyfrom
INSERT INTO player (game_id, user_id, name, turn_index)
VALUES ($1, $2, $3, $4);

