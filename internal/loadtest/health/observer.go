package health

// TrackerObserver adapts a Tracker to the GameObserver interface used by the orchestrator.
type TrackerObserver struct {
	tracker *Tracker
}

// NewTrackerObserver creates an observer that delegates to the given Tracker.
func NewTrackerObserver(t *Tracker) *TrackerObserver {
	return &TrackerObserver{tracker: t}
}

func (o *TrackerObserver) OnGameStarted(gameIndex int) {
	o.tracker.RegisterGame(gameIndex)
}

func (o *TrackerObserver) OnMove(gameIndex int, phase string) {
	o.tracker.RecordMove(gameIndex, phase)
}

func (o *TrackerObserver) OnPhaseChange(gameIndex int, newPhase string) {
	o.tracker.RecordPhaseChange(gameIndex, newPhase)
}

func (o *TrackerObserver) OnGameComplete(gameIndex int) {
	o.tracker.CompleteGame(gameIndex)
}
