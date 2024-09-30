-- name: HasConqueredInTurn :one
select exists
           (select p.id
            from game g
                     join phase p on p.game_id = g.id
            where g.id = $1
              and p.type = 'CONQUER'
              and p.turn = $2);