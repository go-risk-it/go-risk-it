package ctx_test

import (
	"context"
	"log/slog"
	"testing"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	lobbyclx "github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestUserContext_SlogAttrs(t *testing.T) {
	t.Parallel()

	traceCtx := ctx.WithSpan(context.Background(), noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, "user-42")

	attrs := userCtx.SlogAttrs()

	require.Len(t, attrs, 1)
	require.Equal(t, slog.String("user_id", "user-42"), attrs[0])
}

func TestGameContext_SlogAttrs_ComposesUserAndGame(t *testing.T) {
	t.Parallel()

	traceCtx := ctx.WithSpan(context.Background(), noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, "player-1")
	gameCtx := gamectx.WithGameID(userCtx, 42)

	attrs := gameCtx.SlogAttrs()

	require.Len(t, attrs, 2, "GameContext must return both user_id and game_id")

	attrMap := attrsToMap(attrs)
	require.Equal(t, "player-1", attrMap["user_id"].String())
	require.Equal(t, int64(42), attrMap["game_id"].Int64())
}

func TestLobbyContext_SlogAttrs_ComposesUserAndLobby(t *testing.T) {
	t.Parallel()

	traceCtx := ctx.WithSpan(context.Background(), noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, "host-1")
	lobbyCtx := lobbyclx.WithLobbyID(userCtx, 99)

	attrs := lobbyCtx.SlogAttrs()

	require.Len(t, attrs, 2, "LobbyContext must return both user_id and lobby_id")

	attrMap := attrsToMap(attrs)
	require.Equal(t, "host-1", attrMap["user_id"].String())
	require.Equal(t, int64(99), attrMap["lobby_id"].Int64())
}

// attrsToMap converts a slice of slog.Attr to a map keyed by attr name for easy lookup.
func attrsToMap(attrs []slog.Attr) map[string]slog.Value {
	result := make(map[string]slog.Value, len(attrs))
	for _, attr := range attrs {
		result[attr.Key] = attr.Value
	}

	return result
}
