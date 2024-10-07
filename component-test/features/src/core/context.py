import dataclasses

from behave.runner import Context

from src.api.board_state_message import Region
from src.api.card_state_message import CardStateData
from src.api.game_state_message import GameStateData
from src.api.player_state_message import PlayerStateData
from src.client.http_client import RiskItClient
from src.client.supabase_client import SupabaseClient
from src.core.player import Player
from src.core.runner import ServiceRunner


class IndexedBoardStateData:
    def __init__(self, regions: list[Region]):
        self.regions = {region.id: region for region in regions}


class RiskItContext(Context):
    game_id: int
    board_state: IndexedBoardStateData
    card_state: CardStateData
    game_state: GameStateData
    player_state: PlayerStateData
    supabase_client: SupabaseClient
    service_runner: ServiceRunner
    players: dict[str, Player]
    risk_it_clients: dict[str, RiskItClient]
    admin_http_client: RiskItClient
