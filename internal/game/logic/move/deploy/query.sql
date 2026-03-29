-- name: GetDeployableTroops :one
SELECT deploy_phase.deployable_troops
FROM game.game
         JOIN game.phase ON game.current_phase_id = game.phase.id
         JOIN game.deploy_phase ON game.phase.id = game.deploy_phase.phase_id
WHERE game.id = $1;

-- name: DecreaseDeployableTroops :exec
UPDATE game.deploy_phase
SET deployable_troops = game.deploy_phase.deployable_troops - $2
WHERE id = (select dp.id
            from game.game g
                     join game.phase p on g.current_phase_id = p.id
                     join game.deploy_phase dp on p.id = dp.phase_id
            where g.id = $1);