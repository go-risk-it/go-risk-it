import json
import logging

from behave import *
from websockets.sync.client import connect

from src.api.board_state_message import BoardStateMessage
from src.api.card_state_message import CardStateMessage
from src.api.game_state_message import GameStateMessage
from src.api.player_state_message import PlayerStateMessage
from src.api.move_history_message import MoveHistoryMessage
from src.core.context import RiskItContext, IndexedBoardStateData

LOGGER = logging.getLogger(__name__)


@when("{player} connects to the game")
def step_impl(context: RiskItContext, player: str):
    conn = connect(
        f"ws://localhost:8000/ws?gameID={context.game_id}",
        open_timeout=2,
        additional_headers={"Authorization": f"Bearer {context.players[player].user.jwt}"},
    )
    context.players[player].connection = conn


def deserialize(context: RiskItContext, message: str) -> None:
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
        case "cardState":
            card_state_message = CardStateMessage.parse_obj(parsed_message)
            context.card_state = card_state_message.data
        case "moveHistory":
            MoveHistoryMessage.parse_obj(parsed_message)
        case _:
            raise ValueError(f"Unknown message type: {message_type}")


def receive_all_state_updates(context: RiskItContext, player: str):
    conn = context.players[player].connection
    while True:
        try:
            message = conn.recv(timeout=0.001)
            deserialize(context, message)
        except TimeoutError:
            LOGGER.error("Timed out waiting for message")
            break
        except Exception as e:
            LOGGER.error(e)
            break


@then("{player} receives all state updates")
def step_impl(context: RiskItContext, player: str):
    receive_all_state_updates(context, player)


@then("all players receive all state updates")
def all_players_receive_all_state_updates(context: RiskItContext):
    for player in context.players.keys():
        receive_all_state_updates(context, player)
