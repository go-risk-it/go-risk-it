Feature: Creating a game

  Scenario: create game
    Given a game is created with the following players
      | player    |
      | giovanni  |
      | francesco |
      | gabriele  |
      | vasilii   |
    When gabriele deploys 3 troops in siberia