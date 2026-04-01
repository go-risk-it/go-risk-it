package gamestate

import (
	"encoding/json"
	"time"
)

// PhaseType mirrors the server's game.PhaseType.
type PhaseType string

const (
	Cards     PhaseType = "cards"
	Deploy    PhaseType = "deploy"
	Attack    PhaseType = "attack"
	Conquer   PhaseType = "conquer"
	Reinforce PhaseType = "reinforce"
)

// WSMessage is the WS envelope — matches game/api/messaging.Message.
type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"data"`
}

// GameState uses json.RawMessage for phase state since Go generics
// don't round-trip cleanly through JSON. We parse phase state separately.
type GameState struct {
	ID           int64  `json:"id"`
	Turn         int64  `json:"turn"`
	Phase        Phase  `json:"phase"`
	WinnerUserID string `json:"winnerUserId"`
}

type Phase struct {
	Type  PhaseType       `json:"type"`
	State json.RawMessage `json:"state"`
}

type DeployPhaseState struct {
	DeployableTroops int64 `json:"deployableTroops"`
}

type ConquerPhaseState struct {
	AttackingRegionID string `json:"attackingRegionId"`
	DefendingRegionID string `json:"defendingRegionId"`
	MinTroopsToMove   int64  `json:"minTroopsToMove"`
}

// BoardState matches internal/game/api/messaging/board_state.go.
type BoardState struct {
	Regions []Region `json:"regions"`
}

type Region struct {
	ID      string `json:"id"`
	OwnerID string `json:"ownerId"`
	Troops  int64  `json:"troops"`
}

// PlayersState matches internal/game/api/messaging/players_state.go.
type PlayersState struct {
	Players []Player `json:"players"`
}

type PlayerStatus string

const (
	PlayerAlive PlayerStatus = "alive"
	PlayerDead  PlayerStatus = "dead"
)

type ConnectionStatus string

const (
	Connected    ConnectionStatus = "connected"
	Disconnected ConnectionStatus = "disconnected"
)

type Player struct {
	UserID           string           `json:"userId"`
	Name             string           `json:"name"`
	Index            int64            `json:"index"`
	CardCount        int64            `json:"cardCount"`
	Status           PlayerStatus     `json:"status"`
	ConnectionStatus ConnectionStatus `json:"connectionStatus"`
}

// CardState matches internal/game/api/messaging/card_state.go.
type CardState struct {
	Cards []Card `json:"cards"`
}

type CardType string

const (
	Cavalry   CardType = "cavalry"
	Infantry  CardType = "infantry"
	Artillery CardType = "artillery"
	Jolly     CardType = "jolly"
)

type Card struct {
	ID     int64    `json:"id"`
	Type   CardType `json:"type"`
	Region string   `json:"region"`
}

// MoveHistory matches internal/game/api/messaging/move_performed.go.
type MoveHistory struct {
	Moves []MovePerformed `json:"moves"`
}

type MovePerformed struct {
	UserID  string          `json:"userId"`
	Phase   PhaseType       `json:"phase"`
	Move    json.RawMessage `json:"move"`
	Result  json.RawMessage `json:"result"`
	Created time.Time       `json:"created"`
}

// MissionState — only need the type for perf testing.
type MissionState struct {
	Type    string          `json:"type"`
	Details json.RawMessage `json:"details"`
}
