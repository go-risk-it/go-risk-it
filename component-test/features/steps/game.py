from behave import *

from src.api.player_state_message import Player
from src.core.context import RiskItContext
from util.http_assertions import assert_2xx

valid_phases = {'CARDS', 'DEPLOY', 'ATTACK', 'REINFORCE'}


@given('a game is created with the following players')
def step_impl(context: RiskItContext):
    request = {
        "players": [row.get("player") for row in context.table]
    }

    response = context.risk_it_client.create_game(request)

    assert_2xx(response)
    context.game_id = response.json()["gameId"]


@then("the game phase is {phase}")
def step_impl(context: RiskItContext, phase: str):
    if phase not in valid_phases:
        raise ValueError(f"Unknown phase: {phase}")

    assert context.game_state['currentPhase'] == phase


def extract_player(player_name: str, players: list[Player]) -> Player:
    for p in players:
        if p['id'] == player_name:
            return p

    raise Exception(f"Player {player_name} is not in game")


@then("it's {player_name}'s turn")
def step_impl(context: RiskItContext, player_name: str):
    turn = context.game_state['currentTurn']
    players = context.player_state['players']
    player = extract_player(player_name, players)

    assert turn % len(players) == player['index']
