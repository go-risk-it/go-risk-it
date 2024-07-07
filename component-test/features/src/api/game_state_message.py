from dataclasses import dataclass

from dataclasses_json import dataclass_json


@dataclass_json
@dataclass
class DeployPhaseStateData:
    deployableTroops: int


@dataclass
class DeployPhaseStateMessage:
    type: str
    data: DeployPhaseStateData


@dataclass_json
@dataclass
class ConquerPhaseStateData:
    attackingRegionId: int
    defendingRegionId: int
    minTroopsToMove: int


@dataclass_json
@dataclass
class ConquerPhaseStateMessage:
    type: str
    data: ConquerPhaseStateData


@dataclass_json
@dataclass
class GameStateData:
    id: int
    turn: int
    phase: str


@dataclass
class GameStateMessage:
    type: str
    data: GameStateData
