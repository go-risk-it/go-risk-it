package request

import (
	"errors"
	"fmt"
)

type Player struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
}

type CreateGame struct {
	Players []Player `json:"players"`
}

func (r CreateGame) Validate() error {
	if len(r.Players) == 0 {
		return errors.New("players must not be empty")
	}

	for idx, player := range r.Players {
		if player.UserID == "" {
			return fmt.Errorf("players[%d].userId must not be empty", idx)
		}

		if player.Name == "" {
			return fmt.Errorf("players[%d].name must not be empty", idx)
		}
	}

	return nil
}
