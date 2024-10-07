import enum

from pydantic import BaseModel


class CardType(str, enum.Enum):
    CAVALRY = 'cavalry'
    INFANTRY = 'infantry'
    ARTILLERY = 'artillery'
    JOLLY = 'jolly'


class Card(BaseModel):
    type: CardType
    region: str


class CardStateData(BaseModel):
    cards: list[Card]


class CardStateMessage(BaseModel):
    type: str
    data: CardStateData
