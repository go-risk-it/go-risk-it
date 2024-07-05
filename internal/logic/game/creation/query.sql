-- name: InsertGame :one
INSERT INTO game DEFAULT
VALUES
RETURNING *;