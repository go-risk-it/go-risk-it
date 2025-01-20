-- name: GetAvailableCards :many
select c.*
from game g
         join card c on c.game_id = g.id
where g.id = $1
  and c.owner_id is null;

-- name: DrawCard :exec
update card
set owner_id = (select player.id from player where player.user_id = $2 and player.game_id = $3)
where card.id = $1;

-- name: UnlinkCardsFromOwner :exec
UPDATE card
SET owner_id = NULL
WHERE id = ANY (sqlc.arg(cards)::bigint[]);

-- name: GrantRegionTroops :exec
UPDATE region
set troops = troops + $1
where id = ANY (sqlc.arg(regions)::bigint[]);