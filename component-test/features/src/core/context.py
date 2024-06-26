from behave.runner import Context

from src.api.board_state_message import IndexedBoardStateData
from src.api.game_state_message import GameStateData
from src.api.player_state_message import PlayerStateData
from src.client.http_client import RiskItClient
from src.client.supabase_client import SupabaseClient
from src.core.player import Player
from src.core.runner import ServiceRunner


class RiskItContext(Context):
    game_id: int
    board_state: IndexedBoardStateData
    game_state: GameStateData
    player_state: PlayerStateData
    supabase_client: SupabaseClient
    service_runner: ServiceRunner
    players: dict[str, Player]
    risk_it_clients: dict[str, RiskItClient]
    admin_http_client: RiskItClient
