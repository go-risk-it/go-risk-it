package messaging

type Participant struct {
	UserID string `json:"userId"`
}

type LobbyState struct {
	ID           int64         `json:"id"`
	Participants []Participant `json:"participants"`
}
