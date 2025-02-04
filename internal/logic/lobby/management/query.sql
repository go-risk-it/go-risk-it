-- name: GetAvailableLobbies :many
SELECT l.id, l.game_id, COUNT(p.id) as participant_count
FROM lobby l
         LEFT JOIN participant p on l.id = p.lobby_id
WHERE l.game_id IS NULL
GROUP BY l.id;
