import enum
from typing import TypeVar, Generic, Union, Dict, Any

from pydantic import BaseModel, field_validator


class MissionType(str, enum.Enum):
    TWO_CONTINENTS = "twoContinents"
    TWO_CONTINENTS_PLUS_ONE = "twoContinentsPlusOne"
    EIGHTEEN_TERRITORIES_TWO_TROOPS = "eighteenTerritoriesTwoTroops"
    TWENTY_FOUR_TERRITORIES = "twentyFourTerritories"
    ELIMINATE_PLAYER = "eliminatePlayer"


class TwoContinentMission(BaseModel):
    continent1: str
    continent2: str


class TwoContinentsPlusOneMission(BaseModel):
    continent1: str
    continent2: str


class EmptyState(BaseModel):
    pass


class EighteenTerritoriesTwoTroopsMission(EmptyState):
    pass


class TwentyFourTerritoriesMission(EmptyState):
    pass


class EliminatePlayerMission(BaseModel):
    targetUserId: int


# Type variable for mission details
T = TypeVar('T', bound=Union[
    TwoContinentMission,
    TwoContinentsPlusOneMission,
    EighteenTerritoriesTwoTroopsMission,
    TwentyFourTerritoriesMission,
    EliminatePlayerMission
])


class MissionState(BaseModel, Generic[T]):
    type: MissionType
    details: T

    @field_validator('details', mode='before')
    def validate_details(cls, value, info):
        assert 'type' in info.data, f'Missing type field in data {info.data}'
        mission_type = info.data['type']

        if mission_type == MissionType.TWO_CONTINENTS:
            return TwoContinentMission(**value)
        elif mission_type == MissionType.TWO_CONTINENTS_PLUS_ONE:
            return TwoContinentsPlusOneMission(**value)
        elif mission_type == MissionType.EIGHTEEN_TERRITORIES_TWO_TROOPS:
            return EighteenTerritoriesTwoTroopsMission(**value)
        elif mission_type == MissionType.TWENTY_FOUR_TERRITORIES:
            return TwentyFourTerritoriesMission(**value)
        elif mission_type == MissionType.ELIMINATE_PLAYER:
            return EliminatePlayerMission(**value)
        else:
            raise ValueError(f'Unknown mission type: {mission_type}')

class MissionStateMessage(BaseModel, Generic[T]):
    type: str
    data: MissionState[T]
