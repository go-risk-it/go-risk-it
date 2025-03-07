-- name: CreateMoveLog :one
INSERT INTO game.move_log (game_id,
                      player_id,
                      phase,
                      move_data,
                      result)
VALUES ($1,
        (SELECT id FROM game.player WHERE game_id = $1 AND user_id = $2),
        (SELECT p.type
         FROM game.phase p
                  join game.game g on g.current_phase_id = p.id
         WHERE g.id = $1),
        $3,
        $4)
RETURNING *;

-- name: GetMoveLogs :many
SELECT move_log.phase, move_log.move_data, move_log.result, move_log.created, player.user_id
FROM game.move_log
         JOIN game.player ON game.player.id = player_id
WHERE game.move_log.game_id = $1
ORDER BY created DESC
LIMIT sqlc.arg(max_logs)::bigint;