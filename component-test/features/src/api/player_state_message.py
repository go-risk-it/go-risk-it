from pydantic import BaseModel


class Player(BaseModel):
    userId: str
    name: str
    index: int


class PlayerStateData(BaseModel):
    players: list[Player]


class PlayerStateMessage(BaseModel):
    type: str
    data: PlayerStateData
