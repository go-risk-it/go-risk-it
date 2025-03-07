-- name: GetGame :one
SELECT g.id, p.type AS current_phase, p.turn, winner_player.user_id AS winner_user_id
FROM game.game g
         JOIN game.phase p ON g.current_phase_id = p.id
         LEFT JOIN game.player winner_player ON g.winner_player_id = winner_player.id
WHERE g.id = $1;

-- name: GetUserGames :many
SELECT DISTINCT g.id
FROM game.game g
         JOIN game.player p on g.id = p.game_id
WHERE p.user_id = $1
  and g.winner_player_id IS NULL;