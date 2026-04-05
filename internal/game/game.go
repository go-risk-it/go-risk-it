package game

import (
	"context"
	"fmt"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	gameconfig "github.com/go-risk-it/go-risk-it/internal/game/internal/config"
	gamedata "github.com/go-risk-it/go-risk-it/internal/game/internal/data"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/handlers"
	gamelogic "github.com/go-risk-it/go-risk-it/internal/game/internal/logic"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/mission"
	gamerand "github.com/go-risk-it/go-risk-it/internal/game/internal/rand"
	intsnapshot "github.com/go-risk-it/go-risk-it/internal/game/internal/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/web/routes"
	"github.com/go-risk-it/go-risk-it/internal/game/ws"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/fx"
)

var Module = fx.Options(
	gameconfig.Module,
	gamedata.Module,
	gamelogic.Module,
	gamerand.Module,
	handlers.Module,
	intsnapshot.Module,
	routes.Module,
	ws.Module,
	// Adapt ws.Manager to the narrow Gateway interface for the route handler.
	fx.Provide(func(m ws.Manager) ws.Gateway { return m }),
	// Adapt ws.Manager to handler-local interfaces via adapters.
	fx.Provide(newPresenceAdapter),
	fx.Provide(newScopeLifecycleAdapter),
	fx.Provide(newStatePublisherAdapter),
	// Adapt mission.Service to snapshot.MissionQuerier — bridges the type gap
	// without letting game/snapshot import game/logic/mission (arch rule 23).
	fx.Provide(newMissionQuerierAdapter),
)

// presenceAdapter adapts ws.Manager (context-based) to handlers.Presence (ID-based).
type presenceAdapter struct {
	manager ws.Presence
}

func newPresenceAdapter(m ws.Manager) handlers.Presence {
	return &presenceAdapter{manager: m}
}

func (a *presenceAdapter) ConnectedPlayers(gameID int64) []string {
	ctx := minimalGameContext(gameID)

	return a.manager.GetConnectedPlayers(ctx)
}

// scopeLifecycleAdapter adapts ws.Manager to gameapi.ScopeLifecycle.
type scopeLifecycleAdapter struct {
	manager ws.Lifecycle
}

func newScopeLifecycleAdapter(m ws.Manager) gameapi.ScopeLifecycle {
	return &scopeLifecycleAdapter{manager: m}
}

func (a *scopeLifecycleAdapter) RemoveScope(scopeID int64) {
	ctx := minimalGameContext(scopeID)
	a.manager.RemoveGame(ctx)
}

// statePublisherAdapter adapts ws.Writer to gameapi.StatePublisher.
// It serializes a PlayerView into a WS envelope and sends it to a single
// player identified by playerUserID.
type statePublisherAdapter struct {
	writer ws.Writer
}

func newStatePublisherAdapter(m ws.Manager) gameapi.StatePublisher {
	return &statePublisherAdapter{writer: m}
}

func (a *statePublisherAdapter) PublishState(
	ctx context.Context,
	playerUserID string,
	view *snapshot.PlayerView,
) error {
	rawMessage, err := messaging.BuildMessage(messaging.PlayerViewType, view)
	if err != nil {
		return fmt.Errorf("failed to build player view message: %w", err)
	}

	gameCtx, ok := ctx.(gamectx.GameContext)
	if !ok {
		return fmt.Errorf("expected GameContext, got %T", ctx)
	}

	playerCtx := gamectx.WithGameID(
		kernelctx.WithUserID(gameCtx, playerUserID),
		gameCtx.GameID(),
	)

	a.writer.WriteMessage(playerCtx, rawMessage)

	return nil
}

// minimalGameContext creates a GameContext carrying only the gameID, with a
// noop span and empty userID. Used by adapters that bridge ID-based handler
// interfaces to context-based WS manager methods.
func minimalGameContext(gameID int64) gamectx.GameContext {
	return gamectx.WithGameID(
		kernelctx.WithUserID(
			kernelctx.WithSpan(context.Background(), noop.Span{}),
			"",
		),
		gameID,
	)
}

// missionQuerierAdapter wraps mission.Service to satisfy
// snapshot.MissionQuerier by mapping mission-package return types to
// snapshot-local value types.
type missionQuerierAdapter struct {
	svc mission.Service
}

func newMissionQuerierAdapter(svc mission.Service) intsnapshot.MissionQuerier {
	return &missionQuerierAdapter{svc: svc}
}

func (a *missionQuerierAdapter) GetTwoContinentsMission(
	ctx gamectx.GameContext,
	missionID int64,
) (intsnapshot.TwoContinentsResult, error) {
	m, err := a.svc.GetTwoContinentsMission(ctx, missionID)
	if err != nil {
		return intsnapshot.TwoContinentsResult{}, err
	}

	return intsnapshot.TwoContinentsResult{
		Continent1: m.Continent1,
		Continent2: m.Continent2,
	}, nil
}

func (a *missionQuerierAdapter) GetTwoContinentsPlusOneMission(
	ctx gamectx.GameContext,
	missionID int64,
) (intsnapshot.TwoContinentsPlusOneResult, error) {
	m, err := a.svc.GetTwoContinentsPlusOneMission(ctx, missionID)
	if err != nil {
		return intsnapshot.TwoContinentsPlusOneResult{}, err
	}

	return intsnapshot.TwoContinentsPlusOneResult{
		Continent1: m.Continent1,
		Continent2: m.Continent2,
	}, nil
}

func (a *missionQuerierAdapter) GetEliminatePlayerMission(
	ctx gamectx.GameContext,
	missionID int64,
) (string, error) {
	return a.svc.GetEliminatePlayerMission(ctx, missionID)
}
