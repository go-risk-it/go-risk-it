from behave.runner import Context

from src.client.http_client import RiskItClient
from src.client.supabase_client import SupabaseClient
from src.core.player import Player
from src.core.runner import ServiceRunner
from src.game.api.board_state_message import Region
from src.game.api.card_state_message import CardStateData
from src.game.api.game_state_message import GameStateData
from src.game.api.player_state_message import PlayerStateData
from src.lobby.api.lobbies import UserLobbies
from src.lobby.api.lobby_state_message import LobbyStateData


class IndexedBoardStateData:
    def __init__(self, regions: list[Region]):
        self.regions = {region.id: region for region in regions}


class RiskItContext(Context):
    game_id: int
    lobby_id: int
    board_state: IndexedBoardStateData
    card_state: dict[str, CardStateData]
    game_state: GameStateData
    lobby_state: LobbyStateData
    player_state: PlayerStateData
    supabase_client: SupabaseClient
    service_runner: ServiceRunner
    players: dict[str, Player]
    risk_it_clients: dict[str, RiskItClient]
    admin_http_client: RiskItClient
    lobbies: UserLobbies
