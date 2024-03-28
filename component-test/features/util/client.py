from util.prefix_session import PrefixSession


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

    def create_game(self, body):
        return self.__post(
            "/api/1/game",
            body=body,
        )
