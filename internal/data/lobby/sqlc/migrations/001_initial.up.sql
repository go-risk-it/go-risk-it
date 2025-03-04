SET search_path TO lobby;
CREATE TABLE lobby
(
    id       BIGSERIAL PRIMARY KEY,
    owner_id BIGINT,
    game_id  BIGINT UNIQUE
);

CREATE TABLE participant
(
    id       BIGSERIAL PRIMARY KEY,
    lobby_id BIGINT NOT NULL,
    user_id  TEXT   NOT NULL,
    name     TEXT   NOT NULL,
    FOREIGN KEY (lobby_id) REFERENCES lobby (id),
    CONSTRAINT unique_participant_per_lobby UNIQUE (lobby_id, user_id),
    CONSTRAINT unique_name_per_lobby UNIQUE (lobby_id, name)
);

-- Add the foreign key and NOT NULL constraints
ALTER TABLE lobby
    ADD FOREIGN KEY (owner_id) REFERENCES participant (id);