-- name: InsertCards :copyfrom
INSERT INTO card (region_id, game_id, card_type)
VALUES ($1, $2, $3);