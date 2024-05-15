from typing import Dict

from websocket import create_connection, WebSocket

from src.api.subscribe_message import build_subscribe_message


def connect() -> WebSocket:
    return create_connection("ws://localhost:8000/ws", timeout=2)


class RiskItWebsocketManager:
    player_connections: Dict[str, WebSocket]

    def __init__(self):
        self.player_connections = dict()

    def connect_player(self, player: str, game_id: int):
        if player in self.player_connections:
            raise Exception(f"player {player} is already connected")

        conn = connect()
        conn.send(build_subscribe_message(game_id))

        self.player_connections[player] = conn

    def get_conn(self, player: str) -> WebSocket:
        if player not in self.player_connections:
            raise Exception(f"player {player} is not connected")

        return self.player_connections[player]
