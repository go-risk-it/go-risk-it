from requests import Response

from src.client.prefix_session import PrefixSession
from src.core.player import Player


class RiskItClient:
    player: Player

    def __init__(self, player: Player):
        self.player = player
        self.session = PrefixSession(prefix_url="http://localhost:8000")
        self.session.headers.update({"Authorization": f"Bearer {player.jwt}"})

    def __post(self, url: str, body: dict = None, timeout: int = 2):
        return self.session.post(
            url,
            json=body,
            timeout=timeout,
        )

    def create_game(self, body) -> Response:
        return self.__post(
            "/api/v1/game",
            body=body,
        )

    def deploy(self, game_id: int, body) -> Response:
        return self.__post(
            f"/api/v1/game/{game_id}/move/deploy",
            body=body,
        )

    def reset_state(self) -> Response:
        return self.__post(
            "/api/v1/reset",
        )
