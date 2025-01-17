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

