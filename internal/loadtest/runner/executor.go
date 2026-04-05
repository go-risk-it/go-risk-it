package runner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/client"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
)

// ExecutorHandler executes moves via REST and classifies outcomes.
type ExecutorHandler struct {
	gameCtx *GameSession
}

// Register subscribes to EventMoveDecided.
func (h *ExecutorHandler) Register(bus *Bus) {
	bus.On(EventMoveDecided, h.handle)
}

//nolint:funlen // type-switch error classification
func (h *ExecutorHandler) handle(bus *Bus, e Event) {
	evt, ok := e.(MoveDecidedEvent)
	if !ok {
		return
	}

	idx, ok := h.gameCtx.UserIndex[evt.UserID]
	if !ok {
		bus.Emit(MoveFailedEvent{
			Action:  evt.Action,
			Err:     fmt.Errorf("user %s not found in game context", evt.UserID),
			Fatal:   true,
			ErrType: metrics.ErrorTypeExecution,
		})

		return
	}

	rest := h.gameCtx.Players[idx].REST

	// Capture WS versions BEFORE the REST call. The server commits and
	// broadcasts the WS update before the HTTP response reaches us, so
	// the version may already be bumped by the time MoveSucceeded fires.
	preVersions := make([]uint64, len(h.gameCtx.Players))
	for i, p := range h.gameCtx.Players {
		preVersions[i] = p.WS.View().Version()
	}

	t1 := time.Now()
	err := executeAction(h.gameCtx.Ctx, rest, h.gameCtx.GameID, evt.Action)
	t2 := time.Now()

	if err == nil {
		bus.Emit(MoveSucceededEvent{
			Action:      evt.Action,
			RESTLatency: t2.Sub(t1),
			RESTEndTime: t2,
			PreVersions: preVersions,
		})

		return
	}

	// Classify error.
	var conflictErr *client.ConflictError
	if errors.As(err, &conflictErr) {
		bus.Emit(MoveConflictEvent{Action: evt.Action})

		return
	}

	var staleErr *client.StaleStateError
	if errors.As(err, &staleErr) {
		bus.Emit(MoveFailedEvent{
			Action:  evt.Action,
			Err:     err,
			ErrType: metrics.ErrorTypeStaleState,
		})

		return
	}

	var transientErr *client.TransientError
	if errors.As(err, &transientErr) {
		bus.Emit(MoveFailedEvent{
			Action:  evt.Action,
			Err:     err,
			ErrType: metrics.ErrorTypeTransient,
		})

		return
	}

	bus.Emit(MoveFailedEvent{
		Action:  evt.Action,
		Err:     err,
		ErrType: metrics.ErrorTypeExecution,
	})
}

func executeAction(
	ctx context.Context,
	rest RESTClient,
	gameID int64,
	action *player.Action,
) error {
	switch action.Type {
	case player.ActionDeploy:
		a := action.Deploy

		return rest.Deploy(ctx, gameID, client.DeployMove{
			RegionID:      a.RegionID,
			CurrentTroops: a.CurrentTroops,
			DesiredTroops: a.DesiredTroops,
		})
	case player.ActionAttack:
		a := action.Attack

		return rest.Attack(ctx, gameID, client.AttackMove{
			SourceRegionID:  a.SourceRegionID,
			TargetRegionID:  a.TargetRegionID,
			TroopsInSource:  a.TroopsInSource,
			TroopsInTarget:  a.TroopsInTarget,
			AttackingTroops: a.AttackingTroops,
		})
	case player.ActionConquer:
		return rest.Conquer(ctx, gameID, client.ConquerMove{
			Troops: action.Conquer.Troops,
		})
	case player.ActionReinforce:
		a := action.Reinforce

		return rest.Reinforce(ctx, gameID, client.ReinforceMove{
			SourceRegionID: a.SourceRegionID,
			TargetRegionID: a.TargetRegionID,
			TroopsInSource: a.TroopsInSource,
			TroopsInTarget: a.TroopsInTarget,
			MovingTroops:   a.MovingTroops,
		})
	case player.ActionPlayCards:
		combos := make([]client.CardCombination, len(action.Cards.Combinations))
		for i, ids := range action.Cards.Combinations {
			combos[i] = client.CardCombination{CardIDs: ids}
		}

		return rest.PlayCards(ctx, gameID, client.CardsMove{Combinations: combos})
	case player.ActionAdvance:
		return rest.Advance(ctx, gameID, action.Advance.CurrentPhase)
	default:
		return fmt.Errorf("unknown action type: %d", action.Type)
	}
}
