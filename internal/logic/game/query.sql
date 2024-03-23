-- name: InsertGame :one
INSERT INTO game DEFAULT
VALUES
RETURNING id;

-- name: GetGame :one
SELECT *
FROM game
WHERE id = $1;

-- name: SetGamePhase :exec
UPDATE game
SET phase = $2
WHERE id = $1;