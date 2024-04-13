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