-- name: InsertCards :copyfrom
INSERT INTO card (region_id, game_id, card_type)
VALUES ($1, $2, $3);

-- name: GetCardsForPlayer :many
SELECT c.card_type, r.external_reference
FROM game g
         JOIN player p on g.id = p.game_id
         JOIN card c ON c.owner_id = p.id
         LEFT JOIN region r ON c.region_id = r.id
WHERE g.id = $1
  AND p.user_id = $2;