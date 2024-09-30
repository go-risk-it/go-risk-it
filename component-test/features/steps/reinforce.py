from behave import *

from src.core.context import RiskItContext
from util.http_assertions import assert_2xx


@when("{} reinforces from {} to {} with {} troops")
def step_impl(context: RiskItContext, player: str, source_region: str, target_region: str, troops: str):
    source = context.board_state.regions[source_region]
    target = context.board_state.regions[target_region]
    request = {
        "sourceRegionId": source_region,
        "targetRegionId": target_region,
        "troopsInSource": int(source.troops),
        "troopsInTarget": int(target.troops),
        "movingTroops": int(troops)
    }
    response = context.risk_it_clients[player].reinforce(context.game_id, request)
    assert_2xx(response)
