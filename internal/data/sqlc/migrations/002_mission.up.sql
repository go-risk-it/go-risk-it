BEGIN;

CREATE TYPE mission_type AS ENUM (
    'EIGHTEEN_TERRITORIES_TWO_TROOPS',
    'TWENTY_FOUR_TERRITORIES',
    'TWO_CONTINENTS',
    'TWO_CONTINENTS_PLUS_ONE',
    'ELIMINATE_PLAYER');

ALTER TABLE mission
    ADD COLUMN type mission_type NOT NULL DEFAULT 'EIGHTEEN_TERRITORIES_TWO_TROOPS';

CREATE TABLE two_continents_mission
(
    mission_id  BIGINT PRIMARY KEY,
    continent_1 TEXT NOT NULL,
    continent_2 TEXT NOT NULL,
    FOREIGN KEY (mission_id) REFERENCES mission (id)
);

CREATE TABLE two_continents_plus_one_mission
(
    mission_id  BIGINT PRIMARY KEY,
    continent_1 TEXT NOT NULL,
    continent_2 TEXT NOT NULL,
    FOREIGN KEY (mission_id) REFERENCES mission (id)
);

CREATE TABLE eliminate_player_mission
(
    mission_id       BIGINT PRIMARY KEY,
    target_player_id BIGINT NOT NULL,
    FOREIGN KEY (mission_id) REFERENCES mission (id),
    FOREIGN KEY (target_player_id) REFERENCES player (id)
);

CREATE UNIQUE INDEX two_continents_mission_mission_id_idx ON two_continents_mission (mission_id);
CREATE UNIQUE INDEX two_continents_plus_one_mission_mission_id_idx ON two_continents_plus_one_mission (mission_id);
CREATE UNIQUE INDEX eliminate_player_mission_mission_id_idx ON eliminate_player_mission (mission_id);

COMMIT;
