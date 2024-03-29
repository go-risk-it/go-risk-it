from behave import *

from util.context import RiskItContext
from util.http_assertions import assert_2xx


@given('a game is created with the following players')
def step_impl(context: RiskItContext):
    request = {
        "players": [row.get("player") for row in context.table]
    }

    response = context.risk_it_client.create_game(request)

    assert_2xx(response)
    context.game_id = response.json()["gameId"]
