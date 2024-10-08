Feature: Creating a game

  Background:
    Given francesco creates an account
    And gabriele creates an account
    And giovanni creates an account
    And vasilii creates an account

  Scenario: Create game, connect players, attack and conquer a region
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

    And it's francesco's turn
    And there are 3 deployable troops
    And the game phase is deploy

    When francesco deploys 2 troops in western_united_states
    Then all players receive all state updates
    And it's francesco's turn
    And there are 1 deployable troops
    And the game phase is deploy

    When francesco deploys 1 troops in western_united_states
    Then all players receive all state updates
    And it's francesco's turn
    And the game phase is attack

    When francesco attacks from western_united_states to eastern_united_states until conquering
    Then it's francesco's turn
    And the game phase is conquer

    When francesco conquers with 3 troops
    Then it's francesco's turn
    And the game phase is attack

    When francesco advances from phase attack
    Then all players receive all state updates
    And it's francesco's turn
    And the game phase is reinforce

    When francesco reinforces from eastern_united_states to western_united_states with 2 troops
    Then all players receive all state updates
    And it's francesco's turn
    And the game phase is cards
    And francesco has 1 cards