"""Models for the unified playerView WebSocket message.

The server sends a single playerView message containing all game state for a
player, instead of separate gameState/boardState/playerState/cardState/
missionState messages.
"""

from pydantic import BaseModel

from src.game.api.board_state_message import Region
from src.game.api.card_state_message import Card
from src.game.api.game_state_message import Phase
from src.game.api.mission_state_message import MissionState
from src.game.api.player_state_message import Player


class GameMeta(BaseModel):
    id: int
    turn: int
    winnerUserId: str


class PlayerViewData(BaseModel):
    """The data field of a playerView message.

    Maps to the Go snapshot.PlayerView struct:
      game    -> GameMeta
      phase   -> Phase (reused from game_state_message)
      regions -> list[Region]
      players -> list[Player]  (no connectionStatus in new format)
      cards   -> list[Card]
      mission -> MissionState
    """

    game: GameMeta
    phase: Phase
    regions: list[Region]
    players: list[Player]
    cards: list[Card]
    mission: MissionState


class PlayerViewMessage(BaseModel):
    type: str
    data: PlayerViewData
