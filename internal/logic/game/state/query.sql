-- name: GetGame :one
SELECT game.id, phase.type AS current_phase, phase.turn, game.winner_player_id
FROM game
         JOIN phase ON game.current_phase_id = phase.id
WHERE game.id = $1;
