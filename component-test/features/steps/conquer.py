from behave import *

from src.core.context import RiskItContext
from steps.connection import all_players_receive_all_state_updates
from util.http_assertions import assert_2xx


@when("{player} conquers with {troops} troops")
def step_impl(context: RiskItContext, player: str, troops: str):
    request = {
        "troops": int(troops)
    }
    response = context.risk_it_clients[player].conquer(context.game_id, request)

    assert_2xx(response)
    all_players_receive_all_state_updates(context)
