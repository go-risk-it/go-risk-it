ALTER TABLE game.region
    ADD CONSTRAINT check_troops_non_negative CHECK (troops >= 0);

ALTER TABLE game.mission
    ADD CONSTRAINT unique_player_mission UNIQUE (player_id);

ALTER TABLE game.conquer_phase
    ADD CONSTRAINT check_source_target_different CHECK (source_region_id != target_region_id);
