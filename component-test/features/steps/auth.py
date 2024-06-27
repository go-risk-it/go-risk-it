import logging

from behave import *

from src.client.http_client import RiskItClient
from src.core.context import RiskItContext
from src.core.player import Player
from src.core.user import User

LOGGER = logging.getLogger(__name__)


@given("{player} creates an account")
def step_impl(context: RiskItContext, player: str):
    if player in context.players.keys():
        LOGGER.warning(f"Player {player} already exists")

        return

    email = f"{player}@go-risk.it"
    password = "password"

    user = context.supabase_client.get_user(email, password)
    context.players[player] = Player(user=user, name=player)
    context.risk_it_clients[player] = RiskItClient(context.players[player])
