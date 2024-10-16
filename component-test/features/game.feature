Feature: Creating a game

  Background:
    Given francesco creates an account
    And gabriele creates an account
    And giovanni creates an account

  Scenario: Create game, connect players, attack and conquer a region
    Given a game is created with the following players
      | player    |
      | francesco |
      | gabriele  |
      | giovanni  |
    When francesco connects to the game
    And gabriele connects to the game
    And giovanni connects to the game
    Then francesco receives all state updates
    And gabriele receives all state updates
    And giovanni receives all state updates

    And it's francesco's turn
    And there are 3 deployable troops
    And the game phase is deploy

    When francesco deploys 2 troops in ontario
    Then all players receive all state updates
    And it's francesco's turn
    And there are 1 deployable troops
    And the game phase is deploy

    When francesco deploys 1 troops in ontario
    Then all players receive all state updates
    And it's francesco's turn
    And the game phase is attack

    When francesco attacks from ontario to eastern_united_states until conquering
    Then it's francesco's turn
    And the game phase is conquer

    When francesco conquers with 3 troops
    Then it's francesco's turn
    And the game phase is attack

    When francesco advances from phase attack
    Then all players receive all state updates
    And it's francesco's turn
    And the game phase is reinforce

    When francesco reinforces from eastern_united_states to ontario with 2 troops
    Then all players receive all state updates
    And it's gabriele's turn
    And the game phase is cards
    And francesco has 1 cards

    When gabriele advances from phase cards
    And gabriele deploys 4 troops in brazil
    And gabriele attacks from brazil to argentina until conquering
    And gabriele conquers with 3 troops
    And gabriele advances from phase attack
    And gabriele reinforces from argentina to brazil with 2 troops
    Then all players receive all state updates
    And it's giovanni's turn
    And the game phase is cards
    And gabriele has 1 cards

    When giovanni advances from phase cards
    And giovanni deploys 4 troops in siberia
    And giovanni attacks from siberia to ural until conquering
    And giovanni conquers with 4 troops
    And giovanni advances from phase attack
    And giovanni advances from phase reinforce
    Then all players receive all state updates
    And it's francesco's turn
    And the game phase is cards
    And giovanni has 1 cards

    When francesco advances from phase cards
    And francesco deploys 4 troops in eastern_united_states
    And francesco attacks from eastern_united_states to quebec until conquering
    And francesco conquers with 3 troops
    And francesco advances from phase attack
    And francesco advances from phase reinforce
    Then all players receive all state updates
    And it's gabriele's turn
    And the game phase is cards
    And francesco has 2 cards

    When gabriele advances from phase cards
    And gabriele deploys 4 troops in brazil
    And gabriele attacks from brazil to north_africa until conquering
    And gabriele conquers with 9 troops
    And gabriele advances from phase attack
    And gabriele reinforces from peru to venezuela with 2 troops
    Then all players receive all state updates
    And it's giovanni's turn
    And the game phase is cards
    And gabriele has 2 cards

    When giovanni advances from phase cards
    And giovanni deploys 4 troops in egypt
    And giovanni attacks from ural to afghanistan until conquering
    And giovanni conquers with 3 troops
    And giovanni advances from phase attack
    And giovanni reinforces from irkutsk to ural with 1 troops
    Then all players receive all state updates
    And it's francesco's turn
    And the game phase is cards
    And giovanni has 2 cards

    When francesco advances from phase cards




