import json
from typing import Union

from behave import *

from src.api.board_state_message import BoardStateMessage
from src.api.game_state_message import GameStateMessage
from src.api.player_state_message import PlayerStateMessage
from src.core.context import RiskItContext


@when("{player} connects to the game")
def step_impl(context: RiskItContext, player: str):
    context.websocket_manager.connect_player(player, context.game_id)


def deserialize(context: RiskItContext, message: str) -> Union[BoardStateMessage, GameStateMessage, PlayerStateMessage]:
    parsed_message = json.loads(message)
    message_type = parsed_message["type"]

    if message_type == "gameState":
        game_state = GameStateMessage(**parsed_message)
        context.game_state = game_state.data
        return game_state
    elif message_type == "playerState":
        player_state = PlayerStateMessage(**parsed_message)
        context.player_state = player_state.data
        return player_state
    elif message_type == "boardState":
        board_state = BoardStateMessage(**parsed_message)
        context.board_state = board_state.data
        return board_state

    raise ValueError(f"Unknown message type: {message_type}")


def receive_all_state_updates(context: RiskItContext, player):
    conn = context.websocket_manager.get_conn(player)
    for i in range(3):
        deserialize(context, conn.recv())


@then("{player} receives all state updates")
def step_impl(context: RiskItContext, player: str):
    receive_all_state_updates(context, player)


@then("all players receive all state updates")
def step_impl(context: RiskItContext):
    for player in context.websocket_manager.player_connections.keys():
        receive_all_state_updates(context, player)
