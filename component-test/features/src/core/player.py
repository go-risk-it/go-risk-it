from dataclasses import dataclass
from typing import Optional

from websocket import WebSocket


@dataclass
class Player:
    email: str
    password: str
    name: str
    jwt: str
    connection: Optional[WebSocket] = None
