package snapshot

// GameSnapshot is the top-level aggregate for a complete game state read.
// It contains everything needed to render the game UI for all players.
type GameSnapshot struct {
	Game    GameMeta      `json:"game"`
	Phase   Phase         `json:"phase"`
	Regions []RegionState `json:"regions"`
	Players []PlayerState `json:"players"`
}

// GameMeta holds game-level metadata.
type GameMeta struct {
	ID           int64  `json:"id"`
	Turn         int64  `json:"turn"`
	Seq          int64  `json:"seq"`
	WinnerUserID string `json:"winnerUserId"`
}

// RegionState represents a single region on the board.
type RegionState struct {
	ID      string `json:"id"`
	OwnerID string `json:"ownerId"`
	Troops  int64  `json:"troops"`
}

// PlayerState holds publicly visible player information.
type PlayerState struct {
	UserID    string       `json:"userId"`
	Name      string       `json:"name"`
	Index     int64        `json:"index"`
	CardCount int64        `json:"cardCount"`
	Status    PlayerStatus `json:"status"`
}

// PlayerPrivate holds per-player private state (cards and mission).
type PlayerPrivate struct {
	Cards   []CardState   `json:"cards"`
	Mission PlayerMission `json:"mission"`
}

// CardState represents a single card in a player's hand.
type CardState struct {
	ID     int64    `json:"id"`
	Type   CardType `json:"type"`
	Region string   `json:"region"`
}
