package snapshot

// PlayerView is the unified per-player WS message. Each player sees the same
// public state but different cards and mission.
type PlayerView struct {
	Game    GameMeta      `json:"game"`
	Phase   Phase         `json:"phase"`
	Regions []RegionState `json:"regions"`
	Players []PlayerState `json:"players"`
	Cards   []CardState   `json:"cards"`
	Mission PlayerMission `json:"mission"`
}

// BuildPlayerView merges a public snapshot with a player's private data into a
// single PlayerView ready for serialization and dispatch.
func BuildPlayerView(public *GameSnapshot, private *PlayerPrivate) *PlayerView {
	return &PlayerView{
		Game:    public.Game,
		Phase:   public.Phase,
		Regions: public.Regions,
		Players: public.Players,
		Cards:   private.Cards,
		Mission: private.Mission,
	}
}
