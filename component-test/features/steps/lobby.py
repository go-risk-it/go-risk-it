from behave import *
from websockets.sync.client import connect

from src.core.context import RiskItContext
from util.http_assertions import assert_2xx


@given("{player} creates a lobby")
def step_impl(context: RiskItContext, player: str):
    response = context.risk_it_clients[player].create_lobby()

    assert_2xx(response)
    context.lobby_id = response.json()["lobbyId"]


@when("{player} connects to the lobby")
def step_impl(context: RiskItContext, player: str):
    conn = connect(
        f"ws://localhost:8000/ws?lobbyID={context.lobby_id}",
        open_timeout=2,
        additional_headers={"Authorization": f"Bearer {context.players[player].user.jwt}"},
    )
    context.players[player].connection = conn


@when("{player} joins the lobby")
def step_impl(context: RiskItContext, player: str):
    response = context.risk_it_clients[player].join_lobby(context.lobby_id)

    assert_2xx(response)
