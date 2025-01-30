from behave import *
from src.game.api.card_state_message import CardType
from src.core.context import RiskItContext
from steps.connection import all_players_receive_all_state_updates
from util.http_assertions import assert_2xx


@when("{} plays the following card combinations")
def step_impl(context: RiskItContext, player: str):
    def find_unused_card_id_by_type(card_type: str, used: set[int]) -> int:
        card_type_enum = CardType(card_type.lower())
        for card in context.card_state[player].cards:
            if card.type == card_type_enum and card.id not in used:
                used.add(card.id)
                return card.id
        raise ValueError(f"No unused card found with type {card_type}")

    request = {
        "combinations": []
    }

    for row in context.table:
        used_card_ids = set()
        combination = {
            "cardIds": [
                find_unused_card_id_by_type(row["card1"], used_card_ids),
                find_unused_card_id_by_type(row["card2"], used_card_ids),
                find_unused_card_id_by_type(row["card3"], used_card_ids)
            ]
        }
        request["combinations"].append(combination)

    response = context.risk_it_clients[player].cards(context.game_id, request)
    assert_2xx(response)
    all_players_receive_all_state_updates(context)
