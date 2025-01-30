from pydantic import BaseModel


class Region(BaseModel):
    id: str
    ownerId: str
    troops: int


class BoardStateData(BaseModel):
    regions: list[Region]


class BoardStateMessage(BaseModel):
    type: str
    data: BoardStateData
