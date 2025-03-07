-- name: InsertGame :one
INSERT INTO game.game DEFAULT
VALUES
RETURNING *;