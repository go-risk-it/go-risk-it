from behave.runner import Context

from src.api.board_state_message import BoardStateData
from src.api.game_state_message import GameStateData
from src.api.player_state_message import PlayerStateData
from src.client.rest.client import RiskItClient
from src.client.rest.prefix_session import PrefixSession
from src.client.websockets.manager import RiskItWebsocketManager
from src.core.runner import ServiceRunner


class RiskItContext(Context):
    game_id: int
    board_state: BoardStateData
    game_state: GameStateData
    player_state: PlayerStateData
    session: PrefixSession
    websocket_manager: RiskItWebsocketManager
    risk_it_client: RiskItClient
    service_runner: ServiceRunner
