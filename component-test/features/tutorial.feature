Feature: showing off behave

  Scenario: run a simple test
    Given we have behave installed
    When we implement a test
    Then behave will test it for us!

  Scenario: create game
    Given a game is created with the following players
      | player   |
      | giovanni  |
      | francesco |
      | gabriele  |
      | vasilii   |