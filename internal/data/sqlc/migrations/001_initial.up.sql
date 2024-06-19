CREATE TYPE phase AS ENUM ('CARDS', 'DEPLOY', 'ATTACK', 'REINFORCE');

CREATE TABLE game
(
    id                BIGSERIAL PRIMARY KEY,
    turn              BIGINT NOT NULL DEFAULT 0,
    phase             phase  NOT NULL DEFAULT 'DEPLOY',
    deployable_troops BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT check_deployable_troops
        CHECK (
            CASE
                WHEN phase = 'DEPLOY' OR phase = 'CARDS' THEN deployable_troops >= 0
                ELSE deployable_troops = 0
                END
            )
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