-- name: GetPlayersByGame :many
SELECT *
FROM game.player
WHERE game_id = $1;

-- name: GetPlayersState :many
SELECT p.user_id, p.name, p.turn_index, COUNT(distinct c.id) as card_count, COUNT(distinct r.id) as region_count
FROM game.player p
         LEFT JOIN game.card c on p.id = c.owner_id
         LEFT JOIN game.region r on r.player_id = p.id
WHERE p.game_id = $1
GROUP BY p.id
ORDER BY p.turn_index;

-- name: GetPlayerByUserId :one
SELECT *
FROM game.player
WHERE user_id = $1;

-- name: InsertPlayers :copyfrom
INSERT INTO game.player (game_id, user_id, name, turn_index)
VALUES ($1, $2, $3, $4);

-- name: GetNextPlayer :one
SELECT *
FROM game.player
WHERE game.player.game_id = $1
  AND game.player.turn_index = (
    (1 + (SELECT p.turn
          FROM game.game g
                   JOIN game.phase p on g.current_phase_id = p.id
          WHERE g.id = $1))
        % (SELECT COUNT(game.player.id) FROM game.player WHERE game.player.game_id = $1));

-- name: GetPlayerAtTurnIndex :one
SELECT *
FROM game.player
WHERE game.player.game_id = $1
  AND game.player.turn_index = (sqlc.arg(turn) % (SELECT COUNT(game.player.id) FROM game.player WHERE game.player.game_id = $1));

-- name: GetCurrentPlayer :one
SELECT *
FROM game.player
WHERE game.player.game_id = $1
  AND game.player.turn_index = ((SELECT p.turn
                                 FROM game.game g
                                          JOIN game.phase p on g.current_phase_id = p.id
                                 WHERE g.id = $1)
    % (SELECT COUNT(player.id) FROM game.player WHERE player.game_id = $1));


