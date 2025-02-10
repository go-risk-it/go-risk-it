import logging
import time

from dotenv import load_dotenv

from src.client.http_client import RiskItClient
from src.client.supabase_client import SupabaseClient
from src.core.context import RiskItContext
from src.core.player import Player
from src.core.runner import ServiceRunner
from util.readiness import RestReadinessCheck

LOGGER = logging.getLogger(__name__)


def before_all(context: RiskItContext):
    start_command = [
        "docker",
        "compose",
        "up",
        "--build",
        "--detach",
    ]

    load_dotenv()
    context.players = dict()
    context.risk_it_clients = dict()
    context.supabase_client = SupabaseClient()
    context.service_runner = ServiceRunner(
        start_command=start_command,
        path="../",
        timeout=10,
        readiness_check=RestReadinessCheck(["http://localhost:8080/status"]),
    )

    LOGGER.info("Starting service")
    context.service_runner.start()
    LOGGER.info("Service started")

    time.sleep(2)

    setup_admin_account(context)


def setup_admin_account(context: RiskItContext):
    user = context.supabase_client.get_user("admin@admin.admin", "secret_password")

    context.admin_http_client = RiskItClient(Player(user, "admin"))


def before_scenario(context: RiskItContext, _):
    context.admin_http_client.reset_state()


def after_scenario(context: RiskItContext, _):
    close_ws_connections(context)
    close_http_connections(context)


def after_all(context: RiskItContext):
    context.admin_http_client.session.close()


def close_http_connections(context):
    for client in context.risk_it_clients.values():
        client.session.close()
    context.risk_it_clients = dict()


def close_ws_connections(context):
    for player in context.players.values():
        if player.connection is not None:
            player.connection.close()
    context.players = dict()
