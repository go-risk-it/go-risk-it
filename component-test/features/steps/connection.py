import json
import logging

from behave import *
from websocket import create_connection

from src.api.board_state_message import BoardStateMessage
from src.api.game_state_message import GameStateMessage
from src.api.player_state_message import PlayerStateMessage
from src.api.subscribe_message import build_subscribe_message
from src.core.context import RiskItContext, IndexedBoardStateData

LOGGER = logging.getLogger(__name__)


@when("{player} connects to the game")
def step_impl(context: RiskItContext, player: str):
    conn = create_connection(
        "ws://localhost:8000/ws",
        timeout=2,
        header=["Authorization: Bearer " + context.players[player].user.jwt],
    )
    context.players[player].connection = conn
    conn.send(build_subscribe_message(context.game_id))


def deserialize(
        context: RiskItContext, message: str
) -> None:
    parsed_message = json.loads(message)
    message_type = parsed_message["type"]

    LOGGER.info(f"Received message: {message}")

    match message_type:
        case "gameState":
            game_state_message = GameStateMessage.parse_obj(parsed_message)
            context.game_state = game_state_message.data
        case "playerState":
            player_state_message = PlayerStateMessage.parse_obj(parsed_message)
            context.player_state = player_state_message.data
        case "boardState":
            board_state_message = BoardStateMessage.parse_obj(parsed_message)
            context.board_state = IndexedBoardStateData(board_state_message.data.regions)
        case _:
            raise ValueError(f"Unknown message type: {message_type}")


def receive_all_state_updates(context: RiskItContext, player: str):
    conn = context.players[player].connection
    for i in range(3):
        deserialize(context, conn.recv())


@then("{player} receives all state updates")
def step_impl(context: RiskItContext, player: str):
    receive_all_state_updates(context, player)


@then("all players receive all state updates")
def step_impl(context: RiskItContext):
    for player in context.players.keys():
        receive_all_state_updates(context, player)
