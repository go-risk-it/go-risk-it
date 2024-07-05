CREATE TYPE phase_type AS ENUM ('CARDS', 'DEPLOY', 'ATTACK', 'CONQUER', 'REINFORCE');


CREATE TABLE game
(
    id               BIGSERIAL PRIMARY KEY,
    current_phase_id BIGINT
);

CREATE TABLE player
(
    id         BIGSERIAL PRIMARY KEY,
    game_id    BIGINT NOT NULL,
    name       TEXT   NOT NULL,
    user_id    TEXT   NOT NULL,
    turn_index BIGINT NOT NULL CHECK (turn_index >= 0),
    FOREIGN KEY (game_id) REFERENCES game (id),
    CONSTRAINT unique_name_per_game UNIQUE (game_id, name),
    CONSTRAINT unique_user_id_per_game UNIQUE (game_id, user_id),
    CONSTRAINT unique_turn_index_per_game UNIQUE (game_id, turn_index)
);

CREATE TABLE region
(
    id                 BIGSERIAL PRIMARY KEY,
    external_reference TEXT   NOT NULL,
    player_id          BIGINT NOT NULL,
    troops             BIGINT NOT NULL,
    FOREIGN KEY (player_id) REFERENCES player (id)
);

CREATE TABLE card
(
    id        BIGSERIAL PRIMARY KEY,
    player_id BIGINT,
    region_id BIGINT NOT NULL,
    FOREIGN KEY (player_id) REFERENCES player (id),
    FOREIGN KEY (region_id) REFERENCES region (id)
);

CREATE TABLE mission
(
    id        BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL,
    FOREIGN KEY (player_id) REFERENCES player (id)
);

CREATE TABLE phase
(
    id      BIGSERIAL PRIMARY KEY,
    game_id BIGINT     NOT NULL,
    type    phase_type NOT NULL,
    turn    BIGINT     NOT NULL,
    FOREIGN KEY (game_id) REFERENCES game (id),
    CONSTRAINT turn_must_be_positive CHECK (turn >= 0)
);

ALTER TABLE game
    ADD FOREIGN KEY (current_phase_id) REFERENCES phase (id);

CREATE TABLE deploy_phase
(
    id                BIGSERIAL PRIMARY KEY,
    phase_id          BIGINT NOT NULL,
    deployable_troops BIGINT NOT NULL,
    FOREIGN KEY (phase_id) REFERENCES phase (id),
    CONSTRAINT check_deployable_troops CHECK (deployable_troops >= 0)
);

CREATE TABLE conquer_phase
(
    id               BIGSERIAL PRIMARY KEY,
    phase_id         BIGINT NOT NULL,
    source_region_id BIGINT NOT NULL,
    target_region_id BIGINT NOT NULL,
    minimum_troops   BIGINT NOT NULL,
    FOREIGN KEY (phase_id) REFERENCES phase (id),
    FOREIGN KEY (source_region_id) REFERENCES region (id),
    FOREIGN KEY (target_region_id) REFERENCES region (id),
    CONSTRAINT check_minimum_troops CHECK (minimum_troops > 0 AND minimum_troops <= 3)
);