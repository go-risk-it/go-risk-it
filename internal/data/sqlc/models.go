// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package sqlc

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type Phase string

const (
	PhaseCARDS     Phase = "CARDS"
	PhaseDEPLOY    Phase = "DEPLOY"
	PhaseATTACK    Phase = "ATTACK"
	PhaseREINFORCE Phase = "REINFORCE"
)

func (e *Phase) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = Phase(s)
	case string:
		*e = Phase(s)
	default:
		return fmt.Errorf("unsupported scan type for Phase: %T", src)
	}
	return nil
}

type NullPhase struct {
	Phase Phase
	Valid bool // Valid is true if Phase is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullPhase) Scan(value interface{}) error {
	if value == nil {
		ns.Phase, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.Phase.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullPhase) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.Phase), nil
}

type Card struct {
	ID       int64
	PlayerID pgtype.Int8
	RegionID int64
}

type Game struct {
	ID               int64
	Turn             int64
	Phase            Phase
	DeployableTroops int64
}

type Mission struct {
	ID       int64
	PlayerID int64
}

type Player struct {
	ID        int64
	GameID    int64
	Name      string
	UserID    string
	TurnIndex int64
}

type Region struct {
	ID                int64
	ExternalReference string
	PlayerID          int64
	Troops            int64
}
