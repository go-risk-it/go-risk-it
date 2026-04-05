import json
import logging

from behave import *
from websockets.sync.client import connect

from src.game.api.player_view_message import PlayerViewMessage
from src.game.api.game_state_message import GameStateData
from src.game.api.player_state_message import PlayerStateData
from src.game.api.card_state_message import CardStateData
from src.core.context import RiskItContext, IndexedBoardStateData
from src.lobby.api.lobby_state_message import LobbyStateMessage

LOGGER = logging.getLogger(__name__)


@when("{player} connects to the game")
def step_impl(context: RiskItContext, player: str):
    conn = connect(
        f"ws://localhost:8000/api/v1/games/{context.game_id}/ws",
        open_timeout=2,
        additional_headers={"Authorization": f"Bearer {context.players[player].user.jwt}"},
    )
    context.players[player].connection = conn


def _decompose_player_view(context: RiskItContext, parsed_message: dict, player: str) -> None:
    """Decompose a playerView message into the legacy context fields.

    The server sends a single playerView message. To avoid rewriting every step
    definition, we split it back into the individual context fields that steps
    already access: game_state, board_state, player_state, card_state.
    """
    pv = PlayerViewMessage.model_validate(parsed_message)
    data = pv.data

    # game_state: reconstruct GameStateData from game meta + phase
    context.game_state = GameStateData(
        id=data.game.id,
        turn=data.game.turn,
        phase=data.phase,
        winnerUserId=data.game.winnerUserId,
    )

    # board_state: indexed by region id
    context.board_state = IndexedBoardStateData(data.regions)

    # player_state: wrap in PlayerStateData
    context.player_state = PlayerStateData(players=data.players)

    # card_state: per-player, keyed by player name
    if not hasattr(context, "card_state"):
        context.card_state = {}
    context.card_state[player] = CardStateData(cards=data.cards)


def deserialize(context: RiskItContext, message: str, player: str) -> None:
    parsed_message = json.loads(message)
    message_type = parsed_message["type"]

    match message_type:
        case "playerView":
            _decompose_player_view(context, parsed_message, player)
        case "playerConnection":
            pass  # presence notification, no state to store
        case "lobbyState":
            lobby_state_message = LobbyStateMessage.model_validate(parsed_message)
            context.lobby_state = lobby_state_message.data
        case _:
            raise ValueError(f"Unknown message type: {message_type}")


def receive_all_state_updates(context: RiskItContext, player: str):
    conn = context.players[player].connection
    while True:
        try:
            message = conn.recv(timeout=0.01)
            deserialize(context, message, player)
        except TimeoutError:
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
