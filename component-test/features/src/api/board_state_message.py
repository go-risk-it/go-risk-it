from dataclasses import dataclass


@dataclass
class Region:
    id: str
    ownerId: str
    troops: int


@dataclass
class BoardStateData:
    regions: list[Region]


@dataclass
class BoardStateMessage:
    type: str
    data: BoardStateData
