from dataclasses import dataclass

from dataclasses_json import dataclass_json


@dataclass_json
@dataclass
class Player:
    userId: str
    name: str
    index: int


@dataclass_json
@dataclass
class PlayerStateData:
    players: list[Player]


@dataclass
class PlayerStateMessage:
    type: str
    data: PlayerStateData
