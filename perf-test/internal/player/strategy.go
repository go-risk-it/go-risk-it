package player

import (
	"github.com/go-risk-it/go-risk-it/perf-test/internal/gamestate"
)

// ActionType identifies which REST endpoint to call.
type ActionType int

const (
	ActionDeploy ActionType = iota
	ActionAttack
	ActionConquer
	ActionReinforce
	ActionPlayCards
	ActionAdvance
)

// Action represents a decision made by the strategy.
type Action struct {
	Type      ActionType
	Deploy    *DeployAction
	Attack    *AttackAction
	Conquer   *ConquerAction
	Reinforce *ReinforceAction
	Cards     *CardsAction
	Advance   *AdvanceAction
}

type DeployAction struct {
	RegionID      string
	CurrentTroops int64
	DesiredTroops int64
}

type AttackAction struct {
	SourceRegionID  string
	TargetRegionID  string
	TroopsInSource  int64
	TroopsInTarget  int64
	AttackingTroops int64
}

type ConquerAction struct {
	Troops int64
}

type ReinforceAction struct {
	SourceRegionID string
	TargetRegionID string
	TroopsInSource int64
	TroopsInTarget int64
	MovingTroops   int64
}

type CardsAction struct {
	Combinations [][]int64
}

type AdvanceAction struct {
	CurrentPhase string
}

// Strategy is the interface for AI players.
type Strategy interface {
	Name() string
	DecideMove(snap gamestate.ViewSnapshot, userID string) (*Action, error)
}
