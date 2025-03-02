import logging

from behave import *
from websockets.sync.client import connect

from src.core.context import RiskItContext
from src.lobby.api.lobbies import UserLobbies
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


@when("{player} gets the list of available lobbies")
def step_impl(context: RiskItContext, player: str):
    response = context.risk_it_clients[player].get_available_lobbies()

    assert_2xx(response)
    context.lobbies = UserLobbies.model_validate(response.json())


@then("the following lobbies are available")
def step_impl(context: RiskItContext):
    check_lobbies(context.lobbies.owned, [row for row in context.table if row["type"] == "owned"])
    check_lobbies(context.lobbies.joined, [row for row in context.table if row["type"] == "joined"])
    check_lobbies(context.lobbies.joinable,
                  [row for row in context.table if row["type"] == "joinable"])


def check_lobbies(lobbies, expected_lobbies):
    num_participants = sorted([lobby.numberOfParticipants for lobby in lobbies])

    requested_lobbies = sorted([int(row["numberOfParticipants"]) for row in expected_lobbies])

    assert num_participants == requested_lobbies, f"Expected {requested_lobbies}, got {num_participants}"


@when("{player} starts the lobby with {numberOfParticipants} participants")
def step_impl(context: RiskItContext, player: str, numberOfParticipants: str):
    lobbies = [l for l in context.lobbies.owned if l.numberOfParticipants == int(numberOfParticipants)]
    assert len(lobbies) == 1, f"Expected 1 lobby with {numberOfParticipants} participants, got {len(lobbies)}"

    target_lobby = lobbies[0]

    response = context.risk_it_clients[player].start_lobby(target_lobby.id)
    assert_2xx(response)
