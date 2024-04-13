from behave import *

from src.api.player_state_message import Player
from src.core.context import RiskItContext


def extract_player(player_name: str, players: list[Player]) -> Player:
    for p in players:
        if p['id'] == player_name:
            return p

    raise Exception(f"Player {player_name} is not in game")


@then("{player_name} has {deployable_troops} deployable troops")
def step_impl(context: RiskItContext, player_name: str, deployable_troops: int):
    player = extract_player(player_name, context.player_state['players'])

    assert int(player['troopsToDeploy']) == int(deployable_troops)


@then("it's {player_name}'s turn")
def step_impl(context: RiskItContext, player_name: str):
    turn = context.game_state['currentTurn']
    players = context.player_state['players']
    player = extract_player(player_name, players)

    assert turn % len(players) == player['index']
