from dataclasses import dataclass


@dataclass
class GameStateData:
    gameId: int
    currentTurn: int
    currentPhase: str


@dataclass
class GameStateMessage:
    type: str
    data: GameStateData
