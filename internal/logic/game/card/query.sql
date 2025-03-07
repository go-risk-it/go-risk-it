-- name: InsertCards :copyfrom
INSERT INTO game.card (region_id, game_id, card_type)
VALUES ($1, $2, $3);

-- name: GetCardsForPlayer :many
SELECT c.id, c.card_type, r.external_reference as region
FROM game.game g
         JOIN game.player p on g.id = p.game_id
         JOIN game.card c ON c.owner_id = p.id
         LEFT JOIN game.region r ON c.region_id = r.id
WHERE g.id = $1
  AND p.user_id = $2;

-- name: TransferCardsOwnership :exec
UPDATE game.card
SET owner_id = (SELECT id from game.player WHERE player.user_id = sqlc.arg('to')::text AND player.game_id = $1)
WHERE owner_id = sqlc.arg('from');