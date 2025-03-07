from pydantic import BaseModel


class Game(BaseModel):
    id: int


class UserGames(BaseModel):
    games: list[Game]
