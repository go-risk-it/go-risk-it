from dataclasses import dataclass

from dataclasses_json import dataclass_json


@dataclass_json
@dataclass
class GameStateData:
    gameId: int
    currentTurn: int
    currentPhase: str
    deployableTroops: int


@dataclass
class GameStateMessage:
    type: str
    data: GameStateData
