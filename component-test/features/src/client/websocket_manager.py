from websocket import create_connection, WebSocket

from src.api.subscribe_message import build_subscribe_message
from src.core.player import Player


def connect_player(player: Player, game_id: int) -> WebSocket:
    conn = create_connection(
        "ws://localhost:8000/ws",
        timeout=2,
        header=["Authorization: Bearer " + player.jwt],
    )
    conn.send(build_subscribe_message(game_id))

    return conn
