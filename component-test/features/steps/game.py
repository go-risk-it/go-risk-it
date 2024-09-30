from behave import *

from src.core.context import RiskItContext
from util.http_assertions import assert_2xx


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
