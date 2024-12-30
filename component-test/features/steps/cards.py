from behave import *

from src.core.context import RiskItContext
from steps.connection import all_players_receive_all_state_updates
from util.http_assertions import assert_2xx


@when("{} plays the following card combinations")
def step_impl(context: RiskItContext, player: str):
    request = {
        "combinations": [
            {
                "cardIds": [int(row["card1"]), int(row["card2"]), int(row["card3"])]
            } for row in context.table
        ]
    }
    response = context.risk_it_clients[player].cards(context.game_id, request)
    assert_2xx(response)
    all_players_receive_all_state_updates(context)
