from dataclasses import dataclass

from behave import *

from util.context import RiskItContext


@given('a game is created with the following players')
def step_impl(context: RiskItContext):
    request = {
        "players": [row.get("player") for row in context.table]
    }
    print(request)

    context.risk_it_client.create_game(request)
