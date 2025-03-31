-- Drop the second batch of indexes
DROP INDEX IF EXISTS game.idx_card_game_id;
DROP INDEX IF EXISTS game.idx_conquer_phase_source_target;
DROP INDEX IF EXISTS game.idx_phase_game_id_type;
DROP INDEX IF EXISTS game.idx_mission_player_id;
DROP INDEX IF EXISTS game.idx_region_external_reference;

-- Drop the first batch of essential indexes
DROP INDEX IF EXISTS game.idx_card_owner_id;
DROP INDEX IF EXISTS game.idx_move_log_game_id_created;
DROP INDEX IF EXISTS game.idx_phase_game_id;
DROP INDEX IF EXISTS game.idx_region_player_id;
DROP INDEX IF EXISTS game.idx_player_game_id;
