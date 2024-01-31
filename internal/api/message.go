package api

import "encoding/json"

type Type string

const (
	GameStateRequestType  Type = "game_state_request"
	GameStateResponseType Type = "game_state_response"
)

type Message struct {
	Type    Type            `json:"type"`
	Payload json.RawMessage `json:"data"`
}
