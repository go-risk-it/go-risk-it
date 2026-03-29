package request

import "errors"

type JoinLobby struct {
	ParticipantName string `json:"participantName"`
}

func (r JoinLobby) Validate() error {
	if r.ParticipantName == "" {
		return errors.New("participantName must not be empty")
	}

	return nil
}
