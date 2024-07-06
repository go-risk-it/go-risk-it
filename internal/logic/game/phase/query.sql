-- name: InsertPhase :one
INSERT INTO phase (game_id, type, turn)
VALUES ($1, $2, $3)
RETURNING *;

-- name: SetGamePhase :exec
UPDATE game
SET current_phase_id = $2
WHERE id = $1;

-- name: InsertDeployPhase :one
INSERT INTO deploy_phase (phase_id, deployable_troops)
VALUES ($1, $2)
RETURNING *;
