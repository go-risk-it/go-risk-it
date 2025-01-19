-- name: GetGame :one
SELECT game.id, phase.type AS current_phase, phase.turn, winner_player.user_id AS winner_user_id
FROM game
         JOIN phase ON game.current_phase_id = phase.id
         LEFT JOIN player winner_player ON game.winner_player_id = winner_player.id
WHERE game.id = $1;
