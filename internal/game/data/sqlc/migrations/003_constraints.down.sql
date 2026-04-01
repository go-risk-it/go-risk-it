ALTER TABLE game.conquer_phase
    DROP CONSTRAINT IF EXISTS check_source_target_different;

ALTER TABLE game.mission
    DROP CONSTRAINT IF EXISTS unique_player_mission;

ALTER TABLE game.region
    DROP CONSTRAINT IF EXISTS check_troops_non_negative;
