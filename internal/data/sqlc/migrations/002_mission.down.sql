BEGIN;

DROP TABLE IF EXISTS eliminate_player_mission;
DROP TABLE IF EXISTS two_continents_plus_one_mission;
DROP TABLE IF EXISTS two_continents_mission;
ALTER TABLE mission
    DROP COLUMN IF EXISTS type;
DROP TYPE IF EXISTS mission_type;

COMMIT;
