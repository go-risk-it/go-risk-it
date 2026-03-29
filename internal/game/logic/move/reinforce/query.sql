-- name: HasConqueredInTurn :one
select exists
           (select p.id
            from game.game g
                     join game.phase p on p.game_id = g.id
            where g.id = $1
              and p.type = 'CONQUER'
              and p.turn = $2);