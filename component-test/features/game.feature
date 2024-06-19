Feature: Creating a game

  Background:
    Given francesco creates an account
    And gabriele creates an account
    And giovanni creates an account
    And vasilii creates an account

  Scenario: Create game, connect players and make some moves
    Given a game is created with the following players
      | player      |
      | francesco |
      | gabriele  |
      | giovanni  |
      | vasilii   |
    When francesco connects to the game
    And gabriele connects to the game
    And giovanni connects to the game
    And vasilii connects to the game
    Then francesco receives all state updates
    And gabriele receives all state updates
    And giovanni receives all state updates
    And vasilii receives all state updates

    And it's francesco's turn
    And there are 3 deployable troops
    And the game phase is DEPLOY

    When francesco deploys 2 troops in eastern_australia
    Then all players receive all state updates
    And it's francesco's turn
    And there are 1 deployable troops
    And the game phase is DEPLOY

    When francesco deploys 1 troops in ontario
    Then all players receive all state updates
    And it's francesco's turn
    And there are 0 deployable troops
    And the game phase is ATTACK

