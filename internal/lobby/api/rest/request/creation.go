package request

import "errors"

type CreateLobby struct {
	OwnerName string `json:"ownerName"`
}

func (r CreateLobby) Validate() error {
	if r.OwnerName == "" {
		return errors.New("ownerName must not be empty")
	}

	return nil
}
