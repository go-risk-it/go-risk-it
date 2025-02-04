from pydantic import BaseModel


class Lobby(BaseModel):
    id: int
    numberOfParticipants: int


class LobbiesList(BaseModel):
    lobbies: list[Lobby]