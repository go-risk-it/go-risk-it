from behave import *

from src.core.context import RiskItContext
from util.http_assertions import assert_2xx


@when("{player} attacks from {source} to {target} with {troops} troops")
def step_impl(context: RiskItContext, player: str, source: str, target: str, troops: int):
    attacking_region = context.board_state.regions[source]
    defending_region = context.board_state.regions[target]

    request = {
        "sourceRegionId": source,
        "targetRegionId": target,
        "troopsInSource": attacking_region.troops,
        "troopsInTarget": defending_region.troops,
        "attackingTroops": int(troops),
    }
    response = context.risk_it_clients[player].attack(context.game_id, request)

    assert_2xx(response)
