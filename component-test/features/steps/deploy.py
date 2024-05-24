from behave import *

from src.core.context import RiskItContext
from util.http_assertions import assert_2xx


@when("{player} deploys {troops} troops in {region}")
def step_impl(context: RiskItContext, player: str, troops: int, region: str):
    request = {
        "userId": context.players[player].user.id,
        "regionId": region,
        "troops": int(troops),
    }
    response = context.risk_it_clients[player].deploy(context.game_id, request)

    assert_2xx(response)
