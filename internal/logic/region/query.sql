-- name: InsertRegions :copyfrom
INSERT INTO region (external_reference, player_id, troops)
VALUES ($1, $2, $3);
