package message

import "encoding/json"

type RequestType string

const (
	GameStateRequestType  RequestType = "game_state_request"
	BoardStateRequestType RequestType = "board_state_request"
)

type RequestMessage struct {
	Type    RequestType     `json:"type"`
	Payload json.RawMessage `json:"data"`
}
