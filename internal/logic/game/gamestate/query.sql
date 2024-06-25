-- name: InsertGame :one
INSERT INTO game (deployable_troops)
VALUES ($1)
RETURNING *;

-- name: GetGame :one
SELECT *
FROM game
WHERE id = $1;

-- name: SetGamePhase :exec
UPDATE game
SET phase = $2
WHERE id = $1;

-- name: DecreaseDeployableTroops :exec
UPDATE game
SET deployable_troops = deployable_troops - $2
WHERE id = $1;