-- name: GetDeployableTroops :one
SELECT deploy_phase.deployable_troops
FROM game
         JOIN phase ON game.current_phase_id = phase.id
         JOIN deploy_phase ON phase.id = deploy_phase.phase_id
WHERE game.id = $1;

-- name: DecreaseDeployableTroops :exec
UPDATE deploy_phase
SET deployable_troops = deploy_phase.deployable_troops - $2
WHERE id = (select dp.id
            from game g
                     join phase p on g.current_phase_id = p.id
                     join deploy_phase dp on p.id = dp.phase_id
            where g.id = $1);