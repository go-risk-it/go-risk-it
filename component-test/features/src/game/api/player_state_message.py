import enum

from pydantic import BaseModel


class PlayerStatus(str, enum.Enum):
    ALIVE = 'alive'
    DEAD = 'dead'


class ConnectionStatus(str, enum.Enum):
    CONNECTED = 'connected'
    DISCONNECTED = 'disconnected'


class Player(BaseModel):
    userId: str
    name: str
    index: int
    cardCount: int
    status: PlayerStatus
    connectionStatus: ConnectionStatus


class PlayerStateData(BaseModel):
    players: list[Player]


class PlayerStateMessage(BaseModel):
    type: str
    data: PlayerStateData
