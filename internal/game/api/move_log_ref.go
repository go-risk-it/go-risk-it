package game

import "time"

// MoveLogRef carries move log data in event payloads using standard Go types,
// avoiding a dependency on sqlc models or pgtype. Created at the boundary
// where orchestration emits events.
type MoveLogRef struct {
	ID       int64
	GameID   int64
	PlayerID int64
	Phase    GamePhaseType
	MoveData []byte
	Result   []byte
	Created  time.Time
}
