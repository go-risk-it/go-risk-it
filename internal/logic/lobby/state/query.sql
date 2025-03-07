-- name: GetLobby :many
SELECT l.id, p.id as participant_id, p.user_id
FROM lobby.lobby l
         join lobby.participant p on p.lobby_id = l.id
where l.id = $1;