//go:build invariant

package invariant

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
)

// SimulationConfig controls a game simulation run.
type SimulationConfig struct {
	NumPlayers int
	MaxMoves   int
	Seed       uint64
}

// SimulationResult holds the outcome of a simulation run.
type SimulationResult struct {
	GameID    int64
	Winner    string
	MoveCount int
	Completed bool
	FinalSnap *GameSnapshot
}

// RunGame creates a game and plays it to completion,
// checking invariants after every move.
func RunGame(
	tb testing.TB,
	harness *Harness,
	cfg SimulationConfig,
) SimulationResult {
	tb.Helper()

	handle := harness.CreateGame(tb, cfg.NumPlayers)
	gen := NewGenerator(harness, cfg.Seed)

	var prev *GameSnapshot

	for move := range cfg.MaxMoves {
		snap := TakeSnapshot(tb, harness, handle.GameID)
		if snap.IsGameOver() {
			return SimulationResult{
				GameID:    handle.GameID,
				Winner:    snap.Winner,
				MoveCount: move,
				Completed: true,
				FinalSnap: snap,
			}
		}

		playerID := snap.CurrentPlayerUserID()
		gCtx := harness.GameCtx(handle.GameID, playerID)

		executeMove(tb, harness, gen, gCtx, snap)

		newSnap := TakeSnapshot(tb, harness, handle.GameID)
		CheckAll(tb, newSnap, prev)
		prev = newSnap
	}

	finalSnap := TakeSnapshot(tb, harness, handle.GameID)

	return SimulationResult{
		GameID:    handle.GameID,
		MoveCount: cfg.MaxMoves,
		Completed: finalSnap.IsGameOver(),
		FinalSnap: finalSnap,
	}
}

func executeMove(
	tb testing.TB,
	harness *Harness,
	gen *Generator,
	gCtx ctx.GameContext,
	snap *GameSnapshot,
) {
	tb.Helper()

	switch snap.Phase {
	case sqlc.GamePhaseTypeCARDS:
		executeCards(tb, harness, gen, gCtx, snap)
	case sqlc.GamePhaseTypeDEPLOY:
		executeDeploy(tb, harness, gen, gCtx, snap)
	case sqlc.GamePhaseTypeATTACK:
		executeAttack(tb, harness, gen, gCtx, snap)
	case sqlc.GamePhaseTypeCONQUER:
		executeConquer(tb, harness, gen, gCtx)
	case sqlc.GamePhaseTypeREINFORCE:
		executeReinforce(tb, harness, gCtx)
	}
}

func executeCards(
	tb testing.TB,
	harness *Harness,
	gen *Generator,
	gCtx ctx.GameContext,
	snap *GameSnapshot,
) {
	tb.Helper()

	cardsMove := gen.CardsMove(tb, gCtx, snap)

	err := harness.CardsOrchestrator.OrchestrateMove(
		gCtx, cardsMove,
	)
	if err != nil {
		tb.Fatalf("cards move failed: %v", err)
	}
}

func executeDeploy(
	tb testing.TB,
	harness *Harness,
	gen *Generator,
	gCtx ctx.GameContext,
	snap *GameSnapshot,
) {
	tb.Helper()

	deployMove := gen.DeployMove(tb, gCtx, snap)

	err := harness.DeployOrchestrator.OrchestrateMove(
		gCtx, deployMove,
	)
	if err != nil {
		tb.Fatalf("deploy move failed: %v", err)
	}
}

func executeAttack(
	tb testing.TB,
	harness *Harness,
	gen *Generator,
	gCtx ctx.GameContext,
	snap *GameSnapshot,
) {
	tb.Helper()

	attackMove, canAttack := gen.AttackMove(tb, gCtx, snap)
	if !canAttack {
		err := harness.AttackAdvancer.AdvancePhase(gCtx)
		if err != nil {
			tb.Fatalf("attack advance failed: %v", err)
		}

		return
	}

	err := harness.AttackOrchestrator.OrchestrateMove(
		gCtx, attackMove,
	)
	if err != nil {
		tb.Fatalf("attack move failed: %v", err)
	}
}

func executeConquer(
	tb testing.TB,
	harness *Harness,
	gen *Generator,
	gCtx ctx.GameContext,
) {
	tb.Helper()

	conquerMove := gen.ConquerMove(tb, gCtx)

	err := harness.ConquerOrchestrator.OrchestrateMove(
		gCtx, conquerMove,
	)
	if err != nil {
		tb.Fatalf("conquer move failed: %v", err)
	}
}

func executeReinforce(
	tb testing.TB,
	harness *Harness,
	gCtx ctx.GameContext,
) {
	tb.Helper()

	err := harness.ReinforceAdvancer.AdvancePhase(gCtx)
	if err != nil {
		tb.Fatalf("reinforce advance failed: %v", err)
	}
}
