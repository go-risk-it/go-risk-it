from pydantic import BaseModel


class SubscribeData(BaseModel):
    gameId: int


class SubscribeMessage(BaseModel):
    type: str
    data: SubscribeData


def build_subscribe_message(game_id) -> str:
    message = SubscribeMessage(type="subscribe", data=SubscribeData(gameId=game_id))

    return message.model_dump_json()
