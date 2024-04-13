import dataclasses
import json
from dataclasses import dataclass


@dataclass
class SubscribeData:
    gameId: int


@dataclass
class SubscribeMessage:
    type: str
    data: SubscribeData


def build_subscribe_message(game_id) -> str:
    message = SubscribeMessage(
        type="subscribe",
        data=SubscribeData(gameId=game_id)
    )

    return json.dumps(dataclasses.asdict(message))
