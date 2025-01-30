package controller_test

// func TestGameControllerImpl_GetGameState(t *testing.T) {
//	t.Parallel()
//
//	// Initialize dependencies
//	log := zap.NewExample().Sugar()
//	boardService := board.NewService(t)
//	creationService := creation.NewService(t)
//	gameService := gamestate.NewService(t)
//
//	// Initialize the state under test
//	controller := gameController.NewGameController(boardService, creationService, gameService)
//
//	// Set up test data
//	ctx := ctx2.WithGameID(ctx2.WithLog(context.Background(), log), 1)
//	gameID := int64(1)
//
//	// Set up expectations for GetGameState method
//	gameService.
//		EXPECT().
//		GetGameState(ctx).
//		Return(&state.Game{
//			ID:    gameID,
//			Turn:  0,
//			CurrentPhase: sqlc.PhaseTypeCARDS,
//		}, nil)
//
//	// Call the method under test
//	gameState, err := controller.GetGameState(ctx)
//
//	// Assert the result
//	require.NoError(t, err)
//	require.Equal(t, message.GameState{
//		ID:    gameID,
//		Turn:  0,
//		CurrentPhase: message.Cards,
//	}, gameState)
//
//	gameService.AssertExpectations(t)
//}
