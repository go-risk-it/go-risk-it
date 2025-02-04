package messaging

type Lobby struct {
	ID                   int64 `json:"id"`
	NumberOfParticipants int64 `json:"numberOfParticipants"`
}

type Lobbies struct {
	Lobbies []Lobby `json:"lobbies"`
}
