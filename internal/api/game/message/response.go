package message

import "encoding/json"

type ResponseType string

const (
	FullStateResponseType  ResponseType = "full_state_response"
	GameStateResponseType  ResponseType = "game_state_response"
	BoardStateResponseType ResponseType = "board_state_response"
)

type ResponseMessage struct {
	Type    ResponseType    `json:"type"`
	Payload json.RawMessage `json:"data"`
}
