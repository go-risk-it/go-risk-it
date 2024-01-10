CREATE TABLE game
(
    id BIGSERIAL PRIMARY KEY
);

CREATE TABLE player
(
    id      BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL,
    user_id TEXT   NOT NULL,
    FOREIGN KEY (game_id) REFERENCES game (id)
);

CREATE TABLE region
(
    id        BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL,
    troops    INT    NOT NULL,
    FOREIGN KEY (player_id) REFERENCES player (id)
);

CREATE TABLE card
(
    id        BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL,
    FOREIGN KEY (player_id) REFERENCES player (id)
);

CREATE TABLE mission
(
    id        BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL,
    FOREIGN KEY (player_id) REFERENCES player (id)
);