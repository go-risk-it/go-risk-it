from dataclasses import dataclass
from typing import Optional

from websocket import WebSocket

from src.core.user import User


@dataclass
class Player:
    user: User
    name: str
    jwt: str
    connection: Optional[WebSocket] = None
