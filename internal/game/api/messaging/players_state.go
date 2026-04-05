package messaging

// ConnectionStatus indicates whether a player is currently connected via WebSocket.
type ConnectionStatus string

const (
	Connected    ConnectionStatus = "connected"
	Disconnected ConnectionStatus = "disconnected"
)

// PresencePayload is the data field of a playerConnection message, broadcast
// to other connected players when a player connects or disconnects.
type PresencePayload struct {
	UserID string           `json:"userId"`
	Status ConnectionStatus `json:"status"`
}
