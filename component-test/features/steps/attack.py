from behave import *

from src.core.context import RiskItContext
from steps.connection import all_players_receive_all_state_updates
from util.http_assertions import assert_2xx


@when("{player} attacks from {source} to {target} with {troops} troops")
def step_impl(context: RiskItContext, player: str, source: str, target: str, troops: int):
    attacking_region = context.board_state.regions[source]
    defending_region = context.board_state.regions[target]

    response = attack(context, attacking_region, defending_region, player, source, target, troops)

    assert_2xx(response)
    all_players_receive_all_state_updates(context)


@when("{player} attacks from {source} to {target} until conquering")
def step_impl(context: RiskItContext, player: str, source: str, target: str):
    while context.board_state.regions[target].troops > 0:
        attacking_region = context.board_state.regions[source]
        defending_region = context.board_state.regions[target]

        attacking_troops = min(attacking_region.troops - 1, 3)
        if attacking_troops <= 0:
            return
        response = attack(context, attacking_region, defending_region, player, source, target,
                          attacking_troops)

        assert_2xx(response)
        all_players_receive_all_state_updates(context)


def attack(context, attacking_region, defending_region, player, source, target, attacking_troops):
    request = {
        "sourceRegionId": source,
        "targetRegionId": target,
        "troopsInSource": attacking_region.troops,
        "troopsInTarget": defending_region.troops,
        "attackingTroops": int(attacking_troops),
    }
    return context.risk_it_clients[player].attack(context.game_id, request)
