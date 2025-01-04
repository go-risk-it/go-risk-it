package messaging

type ConnectionStatus string

const (
	Connected    ConnectionStatus = "connected"
	Disconnected ConnectionStatus = "disconnected"
)

type PlayerStatus string

const (
	Alive PlayerStatus = "alive"
	Dead  PlayerStatus = "dead"
)

type Player struct {
	UserID           string           `json:"userId"`
	Name             string           `json:"name"`
	Index            int64            `json:"index"`
	CardCount        int64            `json:"cardCount"`
	Status           PlayerStatus     `json:"status"`
	ConnectionStatus ConnectionStatus `json:"connectionStatus"`
}

type PlayersState struct {
	Players []Player `json:"players"`
}
