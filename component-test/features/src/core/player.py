from dataclasses import dataclass
from typing import Optional

from websocket import WebSocket

from src.core.user import User


@dataclass
class Player:
    user: User
    name: str
    connection: Optional[WebSocket] = None
