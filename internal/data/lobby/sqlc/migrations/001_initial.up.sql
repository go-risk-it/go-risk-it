CREATE TABLE lobby
(
    id       BIGSERIAL PRIMARY KEY,
    owner_id BIGINT
);

CREATE TABLE participant
(
    id       BIGSERIAL PRIMARY KEY,
    lobby_id BIGINT NOT NULL,
    user_id  TEXT   NOT NULL,
    FOREIGN KEY (lobby_id) REFERENCES lobby (id)
);

-- Add the foreign key and NOT NULL constraints
ALTER TABLE lobby
    ADD FOREIGN KEY (owner_id) REFERENCES participant (id);