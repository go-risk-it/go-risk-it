-- name: GetAllCardsForGame :many
-- Batch fetch all owned cards for a game, returning player_id for per-player partitioning.
-- Used by SnapshotService.GetPrivateSnapshotsByUser to avoid per-player query amplification.
SELECT c.id, c.card_type, r.external_reference AS region, c.owner_id AS player_id
FROM game.card c
         LEFT JOIN game.region r ON c.region_id = r.id
WHERE c.game_id = $1
  AND c.owner_id IS NOT NULL;

-- name: GetAvailableDeck :many
-- Fetch all unowned cards for a game with region info, for cache warming.
-- Used by getOrWarmPrevState to populate CachedGameState.AvailableDeck.
SELECT c.id, c.card_type, r.external_reference AS region
FROM game.card c
         LEFT JOIN game.region r ON c.region_id = r.id
WHERE c.game_id = $1
  AND c.owner_id IS NULL;

-- name: GetAllMissionsForGame :many
-- Batch fetch all missions for a game, joining through player to filter by game.
-- Used by SnapshotService.GetPrivateSnapshotsByUser for per-player partitioning.
SELECT m.id, m.player_id, m.type
FROM game.mission m
         JOIN game.player p ON m.player_id = p.id
WHERE p.game_id = $1;
