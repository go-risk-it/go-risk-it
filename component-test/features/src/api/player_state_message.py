from dataclasses import dataclass


@dataclass
class Player:
    id: str
    index: int
    troopsToDeploy: int


@dataclass
class PlayerStateData:
    players: list[Player]


@dataclass
class PlayerStateMessage:
    type: str
    data: PlayerStateData
