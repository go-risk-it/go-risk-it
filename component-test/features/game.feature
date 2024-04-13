Feature: Creating a game

  Scenario: create game
    Given a game is created with the following players
      | player    |
      | giovanni  |
      | francesco |
      | gabriele  |
      | vasilii   |
    When gabriele connects to the game
    Then gabriele receives all state updates
    When giovanni connects to the game
    Then giovanni receives all state updates
    When francesco connects to the game
    Then francesco receives all state updates
    When vasilii connects to the game
    Then vasilii receives all state updates

    And it's giovanni's turn
    And giovanni has 5 deployable troops
    And the game phase is DEPLOY

    When giovanni deploys 3 troops in eastern_australia
    Then all players receive all state updates
    And it's giovanni's turn
    And giovanni has 2 deployable troops
    And the game phase is DEPLOY

    When giovanni deploys 2 troops in ontario
    Then all players receive all state updates
    And it's giovanni's turn
    And giovanni has 0 deployable troops
    And the game phase is ATTACK

