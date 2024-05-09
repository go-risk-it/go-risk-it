from behave import *

from src.core.context import RiskItContext
from util.http_assertions import assert_2xx

valid_phases = {'CARDS', 'DEPLOY', 'ATTACK', 'REINFORCE'}


@given('a game is created with the following players')
def step_impl(context: RiskItContext):
    request = {
        "players": [row.get("player") for row in context.table]
    }

    response = context.risk_it_client.create_game(request)

    assert_2xx(response)
    context.game_id = response.json()["gameId"]


@then("the game phase is {phase}")
def step_impl(context: RiskItContext, phase: str):
    if phase not in valid_phases:
        raise ValueError(f"Unknown phase: {phase}")

    assert context.game_state['currentPhase'] == phase
