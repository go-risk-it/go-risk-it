-- name: GetGame :one
SELECT game.id, phase.type AS current_phase, phase.turn
FROM game
         JOIN phase ON game.current_phase_id = phase.id
WHERE game.id = $1;

-- name: UpdateGamePhase :exec
UPDATE game
SET current_phase_id = $1
WHERE id = $2;
