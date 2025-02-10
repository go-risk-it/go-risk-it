package messaging

type Lobby struct {
	ID                   int64 `json:"id"`
	NumberOfParticipants int64 `json:"numberOfParticipants"`
}

type Lobbies struct {
	Owned    []Lobby `json:"owned"`
	Joined   []Lobby `json:"joined"`
	Joinable []Lobby `json:"joinable"`
}
