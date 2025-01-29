-- name: InsertLobby :one
INSERT INTO lobby DEFAULT
VALUES
RETURNING *;