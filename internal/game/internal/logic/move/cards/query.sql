-- name: GetAvailableCards :many
select c.*
from game.game g
         join game.card c on c.game_id = g.id
where g.id = $1
  and c.owner_id is null;

-- name: DrawCard :exec
update game.card
set owner_id = (select game.player.id from game.player where game.player.user_id = $2 and game.player.game_id = $3)
where game.card.id = $1;

-- name: UnlinkCardsFromOwner :exec
UPDATE game.card
SET owner_id = NULL
WHERE id = ANY (sqlc.arg(cards)::bigint[]);

-- name: GrantRegionTroops :exec
UPDATE game.region
set troops = troops + $1
where id = ANY (sqlc.arg(regions)::bigint[]);