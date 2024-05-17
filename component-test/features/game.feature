Feature: Creating a game

  Background:
    Given francesco creates an account
    And gabriele creates an account
    And giovanni creates an account
    And vasilii creates an account
#    And francesco logs in
#    And giovanni logs in
#    And gabriele logs in
#    And vasilii logs in

  Scenario: Create game, connect players and make some moves
    Given a game is created with the following players
      | player    |
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

