from behave import *

from src.core.context import RiskItContext
from steps.connection import all_players_receive_all_state_updates
from util.http_assertions import assert_2xx
from src.game.api.games import UserGames


@given("a game is created with the following players")
def step_impl(context: RiskItContext):
    request = {
        "players": [
            {
                "userId": context.players[row.get("player")].user.id,
                "name": row.get("player"),
            }
            for row in context.table
        ]
    }
    response = context.admin_http_client.create_game(request)

    assert_2xx(response)
    context.game_id = response.json()["gameId"]


@then("the game phase is {phase}")
def step_impl(context: RiskItContext, phase: str):
    assert context.game_state.phase.type == phase, \
        f"Expected phase {phase}, but got {context.game_state.phase.type}"


@when("{player} advances from phase {phase}")
def step_impl(context: RiskItContext, player: str, phase: str):
    request = {
        "currentPhase": phase
    }
    response = context.risk_it_clients[player].advance(context.game_id, request)

    assert_2xx(response)
    all_players_receive_all_state_updates(context)


@then("there is no winner yet")
def step_impl(context: RiskItContext):
    assert context.game_state.winnerUserId == "", \
        f"Expected no winner, but got {context.game_state.winnerUserId}"


@then("the winner is {player}")
def step_impl(context: RiskItContext, player: str):
    assert context.game_state.winnerUserId == context.players[player].user.id, \
        f"Expected {player} to be the winner, but got {context.game_state.winnerUserId}"


@when("{player} gets the list of available games")
def step_impl(context: RiskItContext, player: str):
    response = context.risk_it_clients[player].get_available_games()

    assert_2xx(response)
    context.games = UserGames.model_validate(response.json())


@then("{amount} games are available")
def step_impl(context: RiskItContext, amount: str):
    assert len(context.games.games) == int(amount), f"Expected {amount} games, but got {context.games.games}"
