package messaging

import (
	"encoding/json"
	"fmt"
)

// Type identifies the kind of WebSocket message in the envelope.
type Type string

const (
	GameStateType    Type = "gameState"
	BoardStateType   Type = "boardState"
	PlayerStateType  Type = "playerState"
	CardStateType    Type = "cardState"
	MissionStateType Type = "missionState"
	MoveHistoryType  Type = "moveHistory"
)

// Message is the WebSocket envelope sent to game clients.
type Message struct {
	Type    Type            `json:"type"`
	Payload json.RawMessage `json:"data"`
}

// BuildMessage serializes any payload into a typed WebSocket message envelope.
func BuildMessage(messageType Type, payload any) (json.RawMessage, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %w", err)
	}

	rawMessage, err := json.Marshal(Message{
		Type:    messageType,
		Payload: data,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %w", err)
	}

	return rawMessage, nil
}
