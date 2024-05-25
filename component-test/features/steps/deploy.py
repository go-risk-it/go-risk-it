from behave import *

from src.core.context import RiskItContext
from util.http_assertions import assert_2xx


@when("{player} deploys {troops} troops in {region}")
def step_impl(context: RiskItContext, player: str, troops: int, region: str):
    current_troops = context.board_state.regions[region].troops
    request = {
        "userId": context.players[player].user.id,
        "regionId": region,
        "currentTroops": current_troops,
        "desiredTroops": current_troops + int(troops),
    }
    response = context.risk_it_clients[player].deploy(context.game_id, request)

    assert_2xx(response)
