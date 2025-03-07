-- name: GetConquerPhaseState :one
select source_region.external_reference as source_region,
       target_region.external_reference as target_region,
       cp.minimum_troops
from game.game g
         join game.phase p on g.current_phase_id = p.id
         join game.conquer_phase cp on p.id = cp.phase_id
         join game.region source_region on cp.source_region_id = source_region.id
         join game.region target_region on cp.target_region_id = target_region.id
where g.id = $1;