-- name: GetOwnedLobbies :many
SELECT l.id, l.game_id, COUNT(p.id) AS participant_count
FROM lobby l
         JOIN participant p ON l.id = p.lobby_id
WHERE l.game_id IS NULL
  AND l.owner_id = (SELECT p.id FROM participant p WHERE p.user_id = $1 AND p.lobby_id = l.id)
GROUP BY l.id;

-- name: GetJoinedLobbies :many
WITH joined_lobbies AS (SELECT l.id
                        FROM lobby l
                                 JOIN participant p ON l.id = p.lobby_id
                        WHERE l.game_id IS NULL
                          AND p.user_id = $1)
SELECT l.id, l.game_id, COUNT(p.id) AS participant_count
FROM lobby l
         JOIN participant p ON l.id = p.lobby_id
WHERE l.id IN (SELECT id FROM joined_lobbies)
  AND l.owner_id <> (SELECT p.id FROM participant p WHERE p.user_id = $1 AND p.lobby_id = l.id)
GROUP BY l.id;

-- name: GetJoinableLobbies :many
WITH joined_lobbies AS (SELECT l.id
                        FROM lobby l
                                 JOIN participant p ON l.id = p.lobby_id
                        WHERE l.game_id IS NULL
                          AND p.user_id = $1)
SELECT l.id, l.game_id, COUNT(p.id) AS participant_count
FROM lobby l
         JOIN participant p ON l.id = p.lobby_id
WHERE l.id NOT IN (SELECT id FROM joined_lobbies)
GROUP BY l.id;