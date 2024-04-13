from behave.runner import Context

from src.client.rest.client import RiskItClient
from src.client.rest.prefix_session import PrefixSession
from src.client.websockets.manager import RiskItWebsocketManager
from src.core.runner import ServiceRunner


class RiskItContext(Context):
    game_id: int
    session: PrefixSession
    websocket_manager: RiskItWebsocketManager
    risk_it_client: RiskItClient
    service_runner: ServiceRunner
