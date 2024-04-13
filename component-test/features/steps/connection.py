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


def deserialize(message: str) -> Union[BoardStateMessage, GameStateMessage, PlayerStateMessage]:
    parsed_message = json.loads(message)
    message_type = parsed_message["type"]

    if message_type == "gameState":
        return GameStateMessage(**parsed_message)
    elif message_type == "playerState":
        return PlayerStateMessage(**parsed_message)
    elif message_type == "boardState":
        return BoardStateMessage(**parsed_message)

    raise ValueError(f"Unknown message type: {message_type}")


@then("{player} receives all state updates")
def step_impl(context: RiskItContext, player: str):
    conn = context.websocket_manager.get_conn(player)

    for i in range(3):
        message = deserialize(conn.recv())
        print(message)
