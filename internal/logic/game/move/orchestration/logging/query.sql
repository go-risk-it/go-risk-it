-- name: CreateMoveLog :one
INSERT INTO move_log (game_id,
                      player_id,
                      phase,
                      move_data,
                      result)
VALUES ($1,
        (SELECT id FROM player WHERE game_id = $1 AND user_id = $2),
        (SELECT p.type
         FROM phase p
                  join game g on g.current_phase_id = p.id
         WHERE g.id = $1),
        $3,
        $4)
RETURNING *;
