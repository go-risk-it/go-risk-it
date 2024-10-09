from dataclasses import dataclass
from typing import Optional

from src.core.user import User
from websockets.sync.client import ClientConnection


@dataclass
class Player:
    user: User
    name: str
    connection: Optional[ClientConnection] = None
