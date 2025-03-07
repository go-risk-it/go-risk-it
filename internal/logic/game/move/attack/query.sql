-- name: InsertConquerPhase :one
INSERT INTO game.conquer_phase(phase_id, source_region_id, target_region_id, minimum_troops)
VALUES ($1,
        (select r.id
         from game.game g
                  join game.player p on g.id = p.game_id
                  join game.region r on p.id = r.player_id
         where g.id = $2
           and r.external_reference = $3),
        (select r.id
         from game.game g
                  join game.player p on g.id = p.game_id
                  join game.region r on p.id = r.player_id
         where g.id = $2
           and r.external_reference = $4),
        $5)
RETURNING *;