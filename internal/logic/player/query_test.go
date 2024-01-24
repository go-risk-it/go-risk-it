package player

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tomfran/go-risk-it/internal/db"
)

func TestInsertPlayer(t *testing.T) {
	ctx := context.Background()
	q, err := db.GetQuerier(ctx)

	require.NoError(t, err)

	game1, err := q.InsertGame(ctx)
	require.NoError(t, err)
	game2, err := q.InsertGame(ctx)
	require.NoError(t, err)

	players := []db.InsertPlayersParams{
		{GameID: game1, UserID: "Gabriele"},
		{GameID: game1, UserID: "Giovanni"},
		{GameID: game2, UserID: "Francesco"},
	}

	result, err := q.InsertPlayers(ctx, players)
	require.NoError(t, err)

	require.Equal(t, result, int64(3))
}
