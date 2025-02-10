Feature: Creating a lobby, connecting players, starting a game

  Background:
    Given francesco creates an account
    And gabriele creates an account
    And giovanni creates an account

  Scenario: Create some lobbies and get the list of them
    Given francesco creates a lobby
    When gabriele joins the lobby
    Given giovanni creates a lobby
    Given giovanni creates a lobby
    When francesco joins the lobby
    And gabriele joins the lobby
    When francesco gets the list of available lobbies
    Then the following lobbies are available
      | type     | numberOfParticipants |
      | owned    | 2                    |
      | joined   | 3                    |
      | joinable | 1                    |
    When gabriele gets the list of available lobbies
    Then the following lobbies are available
      | type     | numberOfParticipants |
      | joined   | 2                    |
      | joined   | 3                    |
      | joinable | 1                    |
    When giovanni gets the list of available lobbies
    Then the following lobbies are available
      | type     | numberOfParticipants |
      | owned    | 1                    |
      | owned    | 3                    |
      | joinable | 2                    |

#  Scenario: Create lobby
#    Given francesco creates a lobby
#    When francesco connects to the lobby
#    And gabriele connects to the lobby
#    And giovanni connects to the lobby
#    Then all players receive all state updates
#    When gabriele joins the lobby
#    Then all players receive all state updates
#    When giovanni joins the lobby
#    Then all players receive all state updates



