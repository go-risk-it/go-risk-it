from behave import *

from src.api.player_state_message import Player
from src.core.context import RiskItContext


def extract_player(player_name: str, players: list[Player]) -> Player:
    for p in players:
        if p.name == player_name:
            return p

    raise Exception(f"Player {player_name} is not in game")


@then("There are {deployable_troops} deployable troops")
def step_impl(context: RiskItContext, deployable_troops: str):
    assert context.game_state.deploy_phase.deployableTroops == int(deployable_troops), \
        (f"Expected {deployable_troops} deployable troops, "
         f"but got {context.game_state.deploy_phase.deployableTroops}")


@then("it's {player_name}'s turn")
def step_impl(context: RiskItContext, player_name: str):
    turn = context.game_state.turn
    players = context.player_state.players
    player = extract_player(player_name, players)

    assert turn % len(players) == player.index
