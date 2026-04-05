package snapshot

// Participant represents a player in the lobby.
type Participant struct {
	UserID string `json:"userId"`
}

// LobbySnapshot is a point-in-time view of lobby state.
type LobbySnapshot struct {
	ID           int64         `json:"id"`
	Participants []Participant `json:"participants"`
}
