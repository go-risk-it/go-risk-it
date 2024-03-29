from behave import *

from util.context import RiskItContext
from util.http_assertions import assert_2xx


@when("{player} deploys {troops} troops in {region}")
def step_impl(context: RiskItContext, player: str, troops: int, region: str):
    request = {
        "playerId": player,
        "regionId": region,
        "troops": int(troops),
    }
    response = context.risk_it_client.deploy(context.game_id, request)

    assert_2xx(response)
