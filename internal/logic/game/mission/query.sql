-- name: InsertMission :one
INSERT INTO mission (player_id, type)
VALUES ($1, $2)
RETURNING id;

-- name: InsertTwoContinentsMission :exec
INSERT INTO two_continents_mission (mission_id, continent_1, continent_2)
VALUES ($1, $2, $3);

-- name: InsertTwoContinentsPlusOneMission :exec
INSERT INTO two_continents_plus_one_mission (mission_id, continent_1, continent_2)
VALUES ($1, $2, $3);

-- name: InsertEliminatePlayerMission :exec
INSERT INTO eliminate_player_mission (mission_id, target_player_id)
VALUES ($1, $2);

-- name: GetMission :one
SELECT m.*
FROM mission m
         JOIN player p ON m.player_id = p.id
WHERE p.game_id = $1
  AND p.user_id = $2;

-- name: GetTwoContinentsMission :one
SELECT *
FROM two_continents_mission
WHERE mission_id = $1;

-- name: GetTwoContinentsPlusOneMission :one
SELECT *
FROM two_continents_plus_one_mission
WHERE mission_id = $1;

-- name: GetEliminatePlayerMission :one
SELECT *
FROM eliminate_player_mission
WHERE mission_id = $1;

-- name: ReassignMissions :exec
UPDATE mission
SET type = 'TWENTY_FOUR_TERRITORIES'
WHERE id in (SELECT m.id
             FROM mission m
                      JOIN player p on m.player_id = p.id
                      JOIN eliminate_player_mission em on m.id = em.mission_id
                      JOIN player eliminated_player on em.target_player_id = eliminated_player.id
             WHERE p.game_id = $1
               AND eliminated_player.user_id = sqlc.arg(eliminated_player));

-- name: DeleteSpuriousEliminatePlayerMissions :exec
DELETE
FROM eliminate_player_mission
WHERE mission_id in (SELECT m.id
                     FROM mission m
                              JOIN player p on m.player_id = p.id
                     WHERE p.game_id = $1
                       AND m.type = 'TWENTY_FOUR_TERRITORIES');