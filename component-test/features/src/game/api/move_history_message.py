import enum

from pydantic import BaseModel
from typing import Any
from .game_state_message import PhaseType


class MovePerformed(BaseModel):
    userId: str
    phase: PhaseType
    move: Any
    result: Any
    created: str


class MoveHistoryData(BaseModel):
    moves: list[MovePerformed]


class MoveHistoryMessage(BaseModel):
    type: str
    data: MoveHistoryData
