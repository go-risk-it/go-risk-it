// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package sqlc

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Lobby struct {
	ID      int64
	OwnerID pgtype.Int8
}

type Participant struct {
	ID      int64
	LobbyID int64
	UserID  string
	Name    string
}
