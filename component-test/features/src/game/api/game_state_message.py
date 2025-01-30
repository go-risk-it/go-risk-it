import enum
from typing import TypeVar, Generic

from pydantic import BaseModel, field_validator


class PhaseType(str, enum.Enum):
    CARDS = 'cards'
    DEPLOY = 'deploy'
    ATTACK = 'attack'
    CONQUER = 'conquer'
    REINFORCE = 'reinforce'


class EmptyStateData(BaseModel):
    pass


class DeployPhaseStateData(BaseModel):
    deployableTroops: int


class ConquerPhaseStateData(BaseModel):
    attackingRegionId: str
    defendingRegionId: str
    minTroopsToMove: int


T = TypeVar('T', EmptyStateData, DeployPhaseStateData, ConquerPhaseStateData)


class Phase(BaseModel, Generic[T]):
    type: PhaseType
    state: T

    @field_validator('state', mode='before')
    def validate_state(cls, value, info):
        assert 'type' in info.data, f'Missing type field in data {info.data}'
        phase_type = info.data['type']
        if phase_type == 'deploy':
            return DeployPhaseStateData(**value)
        if phase_type == 'attack':
            return EmptyStateData(**value)
        elif phase_type == 'conquer':
            return ConquerPhaseStateData(**value)
        elif phase_type == 'reinforce':
            return EmptyStateData(**value)
        elif phase_type == 'cards':
            return EmptyStateData(**value)
        else:
            raise ValueError(f'Unknown phase type: {phase_type}')


class GameStateData(BaseModel, Generic[T]):
    id: int
    turn: int
    phase: Phase[T]
    winnerUserId: str

    @property
    def deploy_phase(self) -> DeployPhaseStateData:
        assert self.phase.type == 'deploy' and isinstance(self.phase.state, DeployPhaseStateData), \
            f"Expected deploy phase, but got {self.phase.type}"
        return self.phase.state

    @property
    def conquer_phase(self) -> ConquerPhaseStateData:
        assert self.phase.type == 'conquer' and isinstance(self.phase.state, ConquerPhaseStateData)
        return self.phase.state


class GameStateMessage(BaseModel, Generic[T]):
    type: str
    data: GameStateData[T]
