import logging

from behave import *
from websockets.sync.client import connect

from src.core.context import RiskItContext
from src.lobby.api.lobbies import LobbiesList
from util.http_assertions import assert_2xx

LOGGER = logging.getLogger(__name__)


@given("{player} creates a lobby")
def step_impl(context: RiskItContext, player: str):
    request = {"ownerName": player}
    response = context.risk_it_clients[player].create_lobby(request)

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
    request = {"participantName": player}
    response = context.risk_it_clients[player].join_lobby(context.lobby_id, request)

    assert_2xx(response)


@when("getting for the list of available lobbies")
def step_impl(context: RiskItContext):
    response = context.admin_http_client.get_available_lobbies()

    assert_2xx(response)
    context.lobbies = LobbiesList.model_validate(response.json())



@then("the following lobbies are available")
def step_impl(context: RiskItContext):
    lobbies = context.lobbies
    num_participants = sorted([lobby.numberOfParticipants for lobby in lobbies.lobbies])

    requested_lobbies = sorted([int(row["numberOfParticipants"]) for row in context.table])

    assert num_participants == requested_lobbies, f"Expected {requested_lobbies}, got {num_participants}"
