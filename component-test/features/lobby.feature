Feature: Creating a lobby, connecting players, starting a game

  Background:
    Given francesco creates an account
    And gabriele creates an account
    And giovanni creates an account

  Scenario: Create lobby
    Given francesco creates a lobby
    When gabriele joins the lobby
    And giovanni joins the lobby
