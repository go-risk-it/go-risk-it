from behave import *

from src.core.context import RiskItContext
from util.http_assertions import assert_2xx


@given("{player} creates a lobby")
def step_impl(context: RiskItContext, player: str):
    response = context.risk_it_clients[player].create_lobby()

    assert_2xx(response)
    context.lobby_id = response.json()["lobbyId"]


@when("{player} joins the lobby")
def step_impl(context: RiskItContext, player: str):
    response = context.risk_it_clients[player].join_lobby(context.lobby_id)

    assert_2xx(response)
