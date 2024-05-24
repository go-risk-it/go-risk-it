from dataclasses import dataclass

from dataclasses_json import dataclass_json


@dataclass_json
@dataclass
class Region:
    id: str
    ownerId: str
    troops: int


@dataclass_json
@dataclass
class BoardStateData:
    regions: list[Region]


@dataclass
class BoardStateMessage:
    type: str
    data: BoardStateData
