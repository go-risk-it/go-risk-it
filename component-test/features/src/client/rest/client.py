from requests import Response

from src.client.rest.prefix_session import PrefixSession


class RiskItClient:
    def __init__(self, session: PrefixSession):
        self.session = session

    def __post(
            self,
            url: str,
            body: dict = None,
            timeout: int = 2):
        return self.session.post(
            url,
            json=body,
            timeout=timeout,
        )

    def create_game(self, body) -> Response:
        return self.__post(
            "/api/1/game",
            body=body,
        )

    def deploy(self, game_id: int, body) -> Response:
        return self.__post(
            f"/api/1/game/{game_id}/move/deploy",
            body=body,
        )
