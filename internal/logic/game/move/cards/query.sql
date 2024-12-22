-- name: GetAvailableCards :many
select c.*
from game g
         join card c on c.game_id = g.id
where g.id = $1
  and c.owner_id is null;

-- name: DrawCard :exec
update card
set owner_id = $2
where id = $1;

-- name: UnlinkCardsFromOwner :exec
UPDATE card
SET owner_id = NULL
WHERE id  = ANY(sqlc.arg(cards)::bigint[]);