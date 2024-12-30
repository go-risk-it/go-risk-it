from typing import Optional

from requests import Response

from src.client.prefix_session import PrefixSession
from src.core.player import Player


class RiskItClient:
    player: Player

    def __init__(self, player: Optional[Player] = None):
        self.player = player
        self.session = PrefixSession(prefix_url="http://localhost:8000")
        if player is not None:
            self.session.headers.update({"Authorization": f"Bearer {player.user.jwt}"})

    def __get(self, url: str, timeout: int = 2):
        return self.session.get(
            url,
            timeout=timeout,
        )

    def __post(self, url: str, body: dict = None, timeout: int = 2):
        return self.session.post(
            url,
            json=body,
            timeout=timeout,
        )

    def create_game(self, body) -> Response:
        return self.__post(
            "/api/v1/games",
            body=body,
        )

    def deploy(self, game_id: int, body) -> Response:
        return self.__post(
            f"/api/v1/games/{game_id}/moves/deployments",
            body=body,
        )

    def attack(self, game_id: int, body) -> Response:
        return self.__post(
            f"/api/v1/games/{game_id}/moves/attacks",
            body=body,
        )

    def conquer(self, game_id: int, body) -> Response:
        return self.__post(
            f"/api/v1/games/{game_id}/moves/conquers",
            body=body,
        )

    def advance(self, game_id: int, body) -> Response:
        return self.__post(
            f"/api/v1/games/{game_id}/advancements",
            body=body,
        )

    def reinforce(self, game_id: int, body) -> Response:
        return self.__post(
            f"/api/v1/games/{game_id}/moves/reinforcements",
            body=body,
        )

    def cards(self, game_id: int, body) -> Response:
        return self.__post(
            f"/api/v1/games/{game_id}/moves/cards",
            body=body,
        )

    def is_ready(self) -> bool:
        response = self.__get("/status")

        if response.status_code != 200:
            return False

        response = response.json()

        if response["status"] != "OK":
            return False

        return True

    def reset_state(self) -> Response:
        return self.__post(
            "/api/v1/reset",
        )
