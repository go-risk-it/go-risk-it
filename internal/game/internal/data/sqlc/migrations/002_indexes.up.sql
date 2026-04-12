-- Essential game lookup indexes
CREATE INDEX idx_player_game_id ON game.player(game_id);
CREATE INDEX idx_region_player_id ON game.region(player_id);

-- Phase and game state indexes
CREATE INDEX idx_phase_game_id ON game.phase(game_id, turn);

-- Move log essential index (likely used for game history/replay)
CREATE INDEX idx_move_log_game_id_created ON game.move_log(game_id, created);

-- Card lookup by owner
CREATE INDEX idx_card_owner_id ON game.card(owner_id);

-- Mission lookup by player (mission tracking)
CREATE INDEX idx_mission_player_id ON game.mission(player_id);

-- Phase lookup by type and game (specific game state queries)
CREATE INDEX idx_phase_game_id_type ON game.phase(game_id, type);

-- Conquer phase region indexes (for attack/conquer mechanics)
CREATE INDEX idx_conquer_phase_source_target ON game.conquer_phase(source_region_id, target_region_id);

-- Card lookup by game (card management in a specific game)
CREATE INDEX idx_card_game_id ON game.card(game_id);