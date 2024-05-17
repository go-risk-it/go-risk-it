import random

from behave import *

from src.client.http_client import RiskItClient
from src.core.context import RiskItContext
from src.core.player import Player


@given("{player} creates an account")
def step_impl(context: RiskItContext, player: str):
    email = f"{player}@go-risk.it"
    password = "".join([str(random.randint(0, 9)) for _ in range(10)])

    response = context.supabase_client.sign_up(email, password)
    context.players[player] = Player(
        email=email, password=password, name=player, jwt=response.session.access_token
    )
    context.risk_it_clients[player] = RiskItClient(context.players[player])


#
# @given("{player} logs in")
# def step_impl(context: RiskItContext, player: str):
#     response = context.supabase_client.sign_in(
#         context.players[player].email, context.players[player].password
#     )
#     context.players[player].jwt = response.session.access_token
