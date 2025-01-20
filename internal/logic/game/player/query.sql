-- name: GetPlayersByGame :many
SELECT *
FROM player
WHERE game_id = $1;

-- name: GetPlayersState :many
SELECT p.user_id, p.name, p.turn_index, COUNT(distinct c.id) as card_count, COUNT(distinct r.id) as region_count
FROM player p
         LEFT JOIN card c on p.id = c.owner_id
         LEFT JOIN region r on r.player_id = p.id
WHERE p.game_id = $1
GROUP BY p.id
ORDER BY p.turn_index;

-- name: GetPlayerByUserId :one
SELECT *
FROM player
WHERE user_id = $1;

-- name: InsertPlayers :copyfrom
INSERT INTO player (game_id, user_id, name, turn_index)
VALUES ($1, $2, $3, $4);

-- name: GetNextPlayer :one
SELECT *
FROM player
WHERE player.game_id = $1
  AND player.turn_index = ((1 + (SELECT p.turn
                                 FROM game g
                                          JOIN phase p on g.current_phase_id = p.id
                                 WHERE g.id = $1))
    % (SELECT COUNT (player.id) FROM player WHERE player.game_id = $1));

-- name: GetCurrentPlayer :one
SELECT *
FROM player
WHERE player.game_id = $1
  AND player.turn_index = ((SELECT p.turn
                            FROM game g
                                     JOIN phase p on g.current_phase_id = p.id
                            WHERE g.id = $1)
    % (SELECT COUNT (player.id) FROM player WHERE player.game_id = $1));


