package message

import (
	"encoding/json"
	"fmt"
)

type Type string

const (
	GameState    Type = "gameState"
	BoardState   Type = "boardState"
	PlayerState  Type = "playerState"
	CardState    Type = "cardState"
	MissionState Type = "missionState"
	MoveHistory  Type = "moveHistory"
	LobbyState   Type = "lobbyState"
)

type Message struct {
	Type    Type            `json:"type"`
	Payload json.RawMessage `json:"data"`
}

func BuildMessage(
	messageType Type,
	payload interface{},
) (json.RawMessage, error) {
	var result Message
	result.Type = messageType

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %w", err)
	}

	result.Payload = data

	rawMessage, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %w", err)
	}

	return rawMessage, nil
}
