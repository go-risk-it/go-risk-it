-- name: CreateLobby :one
INSERT INTO lobby DEFAULT
VALUES
RETURNING id;

-- name: InsertParticipant :one
INSERT INTO participant (lobby_id, user_id, name)
VALUES ($1, $2, $3)
RETURNING id;

-- name: UpdateLobbyOwner :exec
UPDATE lobby
SET owner_id = $1
WHERE id = $2;