-- name: InsertConquerPhase :one
INSERT INTO conquer_phase(phase_id, source_region_id, target_region_id, minimum_troops)
VALUES ($1,
        (select r.id
         from game g
                  join player p on g.id = p.game_id
                  join region r on p.id = r.player_id
         where g.id = $2
           and r.external_reference = $3),
        (select r.id
         from game g
                  join player p on g.id = p.game_id
                  join region r on p.id = r.player_id
         where g.id = $2
           and r.external_reference = $4),
        $5)
RETURNING *;

-- name: GetPlayerRegionsFromRegion :one
SELECT p.user_id, COUNT(r.id) as region_count
FROM player p
         JOIN region r on r.player_id = p.id
         JOIN region this_region on this_region.player_id = p.id
WHERE p.game_id = $1
  AND this_region.external_reference = $2
GROUP BY p.user_id;