package gamestate

import "encoding/json"

// WSMessage is the WS envelope — matches game/api/messaging.Message.
type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"data"`
}
