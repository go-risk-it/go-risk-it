Feature: Creating a game

  Background:
    Given francesco creates an account
    And gabriele creates an account
    And giovanni creates an account

#  Scenario: Create game, connect players, attack and conquer a region
#    Given a game is created with the following players
#      | player    |
#      | francesco |
#      | gabriele  |
#      | giovanni  |
#    When francesco connects to the game
#    And gabriele connects to the game
#    And giovanni connects to the game
#    Then francesco receives all state updates
#    And gabriele receives all state updates
#    And giovanni receives all state updates
#
#    And it's francesco's turn
#    And there are 3 deployable troops
#    And the game phase is deploy
#
#    When francesco deploys 2 troops in ontario
#    Then all players receive all state updates
#    And it's francesco's turn
#    And there are 1 deployable troops
#    And the game phase is deploy
#
#    When francesco deploys 1 troops in ontario
#    Then all players receive all state updates
#    And it's francesco's turn
#    And the game phase is attack
#
#    When francesco attacks from ontario to eastern_united_states until conquering
#    Then it's francesco's turn
#    And the game phase is conquer
#
#    When francesco conquers with 3 troops
#    Then it's francesco's turn
#    And the game phase is attack
#
#    When francesco advances from phase attack
#    Then all players receive all state updates
#    And it's francesco's turn
#    And the game phase is reinforce
#
#    When francesco reinforces from eastern_united_states to ontario with 2 troops
#    Then all players receive all state updates
#    And it's gabriele's turn
#    And the game phase is deploy
#    And francesco has 1 cards
#
#    When gabriele deploys 4 troops in brazil
#    And gabriele attacks from brazil to argentina until conquering
#    And gabriele conquers with 3 troops
#    And gabriele advances from phase attack
#    And gabriele reinforces from argentina to brazil with 2 troops
#    Then all players receive all state updates
#    And it's giovanni's turn
#    And the game phase is deploy
#    And gabriele has 1 cards
#
#    When giovanni deploys 4 troops in siberia
#    And giovanni attacks from siberia to ural until conquering
#    And giovanni conquers with 4 troops
#    And giovanni advances from phase attack
#    And giovanni advances from phase reinforce
#    Then all players receive all state updates
#    And it's francesco's turn
#    And the game phase is deploy
#    And giovanni has 1 cards
#
#    When francesco deploys 4 troops in eastern_united_states
#    And francesco attacks from eastern_united_states to quebec until conquering
#    And francesco conquers with 3 troops
#    And francesco advances from phase attack
#    And francesco advances from phase reinforce
#    Then all players receive all state updates
#    And it's gabriele's turn
#    And the game phase is deploy
#    And francesco has 2 cards
#
#    When gabriele deploys 6 troops in brazil
#    And gabriele attacks from brazil to north_africa until conquering
#    And gabriele conquers with 9 troops
#    And gabriele advances from phase attack
#    And gabriele reinforces from peru to venezuela with 2 troops
#    Then all players receive all state updates
#    And it's giovanni's turn
#    And the game phase is deploy
#    And gabriele has 2 cards
#
#    When giovanni deploys 4 troops in egypt
#    And giovanni attacks from ural to afghanistan until conquering
#    And giovanni conquers with 3 troops
#    And giovanni advances from phase attack
#    And giovanni reinforces from irkutsk to ural with 1 troops
#    Then all players receive all state updates
#    And it's francesco's turn
#    And the game phase is deploy
#    And giovanni has 2 cards
#
#    When francesco deploys 4 troops in china
#    And francesco attacks from china to siam until conquering
#    And francesco conquers with 5 troops
#    And francesco advances from phase attack
#    And francesco reinforces from india to siam with 2 troops
#    Then all players receive all state updates
#    And it's gabriele's turn
#    And the game phase is deploy
#    And francesco has 3 cards
#
#    When gabriele deploys 6 troops in congo
#    And gabriele attacks from congo to south_africa until conquering
#    And gabriele conquers with 4 troops
#    And gabriele advances from phase attack
#    And gabriele reinforces from north_africa to south_africa with 2 troops
#    Then all players receive all state updates
#    And it's giovanni's turn
#    And the game phase is deploy
#    And gabriele has 3 cards
#
#    When giovanni deploys 5 troops in new_guinea
#    And giovanni attacks from new_guinea to eastern_australia until conquering
#    And giovanni conquers with 4 troops
#    And giovanni advances from phase attack
#    And giovanni reinforces from eastern_australia to western_australia with 3 troops
#    Then all players receive all state updates
#    And it's francesco's turn
#    And the game phase is cards
#    And giovanni has 3 cards
#
#    When francesco plays the following card combinations
#      | card1     | card2    | card3   |
#      | ARTILLERY | INFANTRY | CAVALRY |
#    When francesco deploys 10 troops in middle_east
#    And francesco deploys 3 troops in siam
#    And francesco attacks from siam to indonesia until conquering
#    And francesco conquers with 9 troops
#    And francesco attacks from indonesia to western_australia until conquering
#    And francesco conquers with 8 troops
#    And francesco attacks from western_australia to eastern_australia until conquering
#    And francesco conquers with 7 troops
#    And francesco attacks from eastern_australia to new_guinea until conquering
#    And francesco conquers with 6 troops
#    And francesco attacks from middle_east to ukraine until conquering
#    And francesco conquers with 12 troops
#    And francesco attacks from ukraine to afghanistan until conquering
#    And francesco conquers with 11 troops
#    And francesco attacks from afghanistan to ural until conquering
#    And francesco conquers with 10 troops
#    And francesco attacks from ural to siberia until conquering
#    And francesco conquers with 9 troops
#    And francesco attacks from siberia to irkutsk until conquering
#    And francesco conquers with 8 troops
#    And francesco attacks from irkutsk to yakutsk until conquering
#    And francesco conquers with 7 troops
#    And francesco attacks from scandinavia to iceland until conquering
#    And francesco conquers with 2 troops
#    And francesco advances from phase attack
#    And francesco advances from phase reinforce
#
#    Then it's gabriele's turn
#    And the game phase is deploy
#    When gabriele deploys 5 troops in south_africa
#    And gabriele deploys 1 troops in alaska
#    And gabriele attacks from south_africa to madagascar until conquering
#    And gabriele conquers with 10 troops
#    And gabriele attacks from madagascar to east_africa until conquering
#    And gabriele conquers with 9 troops
#    And gabriele attacks from east_africa to egypt until conquering
#    And gabriele conquers with 8 troops
#    And gabriele attacks from venezuela to central_america until conquering
#    And gabriele conquers with 4 troops
#    And gabriele attacks from alaska to northwest_territory until conquering
#    And gabriele conquers with 3 troops
#    And gabriele attacks from northwest_territory to alberta until conquering
#    And gabriele conquers with 2 troops
#
#    Then giovanni is dead
#    And gabriele has 6 cards
#    When gabriele advances from phase attack
#    And gabriele advances from phase reinforce
#    Then gabriele has 7 cards
#
#    And it's francesco's turn
#    And there is no winner yet









