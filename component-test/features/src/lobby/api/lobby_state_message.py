from pydantic import BaseModel


class Participant(BaseModel):
    userId: str


class LobbyStateData(BaseModel):
    id: int
    participants: list[Participant]


class LobbyStateMessage(BaseModel):
    type: str
    data: LobbyStateData
