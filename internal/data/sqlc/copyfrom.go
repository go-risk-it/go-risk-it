// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: copyfrom.go

package sqlc

import (
	"context"
)

// iteratorForInsertPlayers implements pgx.CopyFromSource.
type iteratorForInsertPlayers struct {
	rows                 []InsertPlayersParams
	skippedFirstNextCall bool
}

func (r *iteratorForInsertPlayers) Next() bool {
	if len(r.rows) == 0 {
		return false
	}
	if !r.skippedFirstNextCall {
		r.skippedFirstNextCall = true
		return true
	}
	r.rows = r.rows[1:]
	return len(r.rows) > 0
}

func (r iteratorForInsertPlayers) Values() ([]interface{}, error) {
	return []interface{}{
		r.rows[0].GameID,
		r.rows[0].UserID,
		r.rows[0].TurnIndex,
	}, nil
}

func (r iteratorForInsertPlayers) Err() error {
	return nil
}

func (q *Queries) InsertPlayers(ctx context.Context, arg []InsertPlayersParams) (int64, error) {
	return q.db.CopyFrom(ctx, []string{"player"}, []string{"game_id", "user_id", "turn_index"}, &iteratorForInsertPlayers{rows: arg})
}

// iteratorForInsertRegions implements pgx.CopyFromSource.
type iteratorForInsertRegions struct {
	rows                 []InsertRegionsParams
	skippedFirstNextCall bool
}

func (r *iteratorForInsertRegions) Next() bool {
	if len(r.rows) == 0 {
		return false
	}
	if !r.skippedFirstNextCall {
		r.skippedFirstNextCall = true
		return true
	}
	r.rows = r.rows[1:]
	return len(r.rows) > 0
}

func (r iteratorForInsertRegions) Values() ([]interface{}, error) {
	return []interface{}{
		r.rows[0].ExternalReference,
		r.rows[0].PlayerID,
		r.rows[0].Troops,
	}, nil
}

func (r iteratorForInsertRegions) Err() error {
	return nil
}

func (q *Queries) InsertRegions(ctx context.Context, arg []InsertRegionsParams) (int64, error) {
	return q.db.CopyFrom(ctx, []string{"region"}, []string{"external_reference", "player_id", "troops"}, &iteratorForInsertRegions{rows: arg})
}
