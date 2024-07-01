-- name: InsertGame :one
INSERT INTO game DEFAULT
VALUES
RETURNING *;

-- name: GetGame :one
SELECT *
FROM game
WHERE id = $1;

-- name: SetGamePhase :exec
UPDATE game
SET current_phase_id = $2
WHERE id = $1;

-- name: DecreaseDeployableTroops :exec
UPDATE deploy_phase
SET deploy_phase.deployable_troops = deploy_phase.deployable_troops - $2
WHERE deploy_phase.id = (select dp.id
                         from game g
                                  join phase p on g.current_phase_id = p.id
                                  join deploy_phase dp on p.id = dp.phase_id
                         where g.id = $1);