from behave import *

from src.client.http_client import RiskItClient
from src.core.context import RiskItContext
from src.core.player import Player
from src.core.user import User


@given("{player} creates an account")
def step_impl(context: RiskItContext, player: str):
    email = f"{player}@go-risk.it"
    password = "password"

    response = context.supabase_client.sign_up(email, password)
    user = User(id=response.user.id, email=email, password=password)
    context.players[player] = Player(
        user=user, name=player, jwt=response.session.access_token
    )
    context.risk_it_clients[player] = RiskItClient(context.players[player])
