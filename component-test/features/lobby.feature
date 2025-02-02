Feature: Creating a lobby, connecting players, starting a game

  Background:
    Given francesco creates an account
    And gabriele creates an account
    And giovanni creates an account

  Scenario: Create lobby
    Given francesco creates a lobby
    When francesco connects to the lobby
    And gabriele connects to the lobby
    And giovanni connects to the lobby
    Then all players receive all state updates
    When gabriele joins the lobby
    Then all players receive all state updates
    When giovanni joins the lobby
    Then all players receive all state updates
