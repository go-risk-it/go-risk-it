package player_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tomfran/go-risk-it/internal/db"
)

func TestInsertPlayer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	querier, err := db.GetQuerier(ctx)

	require.NoError(t, err)

	game1, err := querier.InsertGame(ctx)
	require.NoError(t, err)
	game2, err := querier.InsertGame(ctx)
	require.NoError(t, err)

	players := []db.InsertPlayersParams{
		{GameID: game1, TurnIndex: 1, UserID: "Gabriele"},
		{GameID: game1, TurnIndex: 2, UserID: "Giovanni"},
		{GameID: game2, TurnIndex: 1, UserID: "Francesco"},
	}

	result, err := querier.InsertPlayers(ctx, players)
	require.NoError(t, err)

	require.Equal(t, int64(3), result)
}
