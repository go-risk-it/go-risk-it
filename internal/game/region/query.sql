-- name: InsertRegions :copyfrom
INSERT INTO region (player_id, troops) VALUES ($1, $2);
