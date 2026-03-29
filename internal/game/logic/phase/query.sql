-- name: InsertPhase :one
INSERT INTO game.phase (game_id, type, turn)
VALUES ($1, $2, $3) RETURNING *;

-- name: SetGamePhase :exec
UPDATE game.game
SET current_phase_id = $2
WHERE id = $1;

-- name: InsertDeployPhase :one
INSERT INTO game.deploy_phase (phase_id, deployable_troops)
VALUES ($1, $2) RETURNING *;

-- name: GetCurrentPhase :one
SELECT p.type
FROM game.phase p
         JOIN game.game g on g.current_phase_id = p.id
WHERE g.id = $1;
