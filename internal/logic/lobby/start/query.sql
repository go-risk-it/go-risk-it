-- name: CanLobbyBeStarted :one
SELECT EXISTS(SELECT l.id
              FROM lobby.lobby l
                       JOIN lobby.participant p ON l.id = p.lobby_id
              WHERE l.id = sqlc.arg(lobby_id)
                AND l.game_id IS NULL
                AND l.owner_id =
                    (SELECT p.id FROM lobby.participant p WHERE p.user_id = sqlc.arg(user_id) AND p.lobby_id = l.id)
              GROUP BY l.id
              HAVING COUNT(p.id) >= sqlc.arg(minimum_participants));

-- name: GetLobbyPlayers :many
SELECT p.user_id, p.name
FROM lobby.lobby l
         JOIN lobby.participant p on l.id = p.lobby_id
where l.id = $1;

-- name: MarkLobbyAsStarted :exec
UPDATE lobby.lobby
SET game_id = sqlc.arg(game_id)
WHERE id = sqlc.arg(lobby_id);