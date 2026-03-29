-- name: InsertMission :one
INSERT INTO game.mission (player_id, type)
VALUES ($1, $2)
RETURNING id;

-- name: InsertTwoContinentsMission :exec
INSERT INTO game.two_continents_mission (mission_id, continent_1, continent_2)
VALUES ($1, $2, $3);

-- name: InsertTwoContinentsPlusOneMission :exec
INSERT INTO game.two_continents_plus_one_mission (mission_id, continent_1, continent_2)
VALUES ($1, $2, $3);

-- name: InsertEliminatePlayerMission :exec
INSERT INTO game.eliminate_player_mission (mission_id, target_player_id)
VALUES ($1, $2);

-- name: GetMission :one
SELECT m.*
FROM game.mission m
         JOIN game.player p ON m.player_id = p.id
WHERE p.game_id = $1
  AND p.user_id = $2;

-- name: GetTwoContinentsMission :one
SELECT *
FROM game.two_continents_mission
WHERE mission_id = $1;

-- name: GetTwoContinentsPlusOneMission :one
SELECT *
FROM game.two_continents_plus_one_mission
WHERE mission_id = $1;

-- name: GetEliminatePlayerMission :one
SELECT *
FROM game.eliminate_player_mission
WHERE mission_id = $1;

-- name: GetPlayerToEliminate :one
SELECT p.user_id
FROM game.eliminate_player_mission em
         JOIN game.player p on em.target_player_id = p.id
WHERE mission_id = $1;

-- name: ReassignMissions :exec
UPDATE game.mission
SET type = 'TWENTY_FOUR_TERRITORIES'
WHERE id in (SELECT m.id
             FROM game.mission m
                      JOIN game.player p on m.player_id = p.id
                      JOIN game.eliminate_player_mission em on m.id = em.mission_id
             WHERE p.game_id = $1
               AND em.target_player_id = sqlc.arg(eliminated_player_id)
               AND p.user_id <> $2);

-- name: DeleteSpuriousEliminatePlayerMissions :exec
DELETE
FROM game.eliminate_player_mission
WHERE mission_id in (SELECT m.id
                     FROM game.mission m
                              JOIN game.player p on m.player_id = p.id
                     WHERE p.game_id = $1
                       AND m.type = 'TWENTY_FOUR_TERRITORIES');

-- name: AssignGameWinner :exec
UPDATE game.game
SET winner_player_id = $1
WHERE id = sqlc.arg(game_id);