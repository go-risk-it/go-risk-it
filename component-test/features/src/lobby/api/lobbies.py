from pydantic import BaseModel


class Lobby(BaseModel):
    id: int
    numberOfParticipants: int


class UserLobbies(BaseModel):
    owned: list[Lobby]
    joined: list[Lobby]
    joinable: list[Lobby]
