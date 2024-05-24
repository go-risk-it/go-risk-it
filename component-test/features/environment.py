import logging
import time

from dotenv import load_dotenv

from src.client.http_client import RiskItClient
from src.client.supabase_client import SupabaseClient
from src.core.context import RiskItContext
from src.core.player import Player
from src.core.runner import ServiceRunner
from src.core.user import User

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
    )

    LOGGER.info("Starting service")
    context.service_runner.start()
    LOGGER.info("Service started")

    time.sleep(2)

    setup_admin_account(context)


def setup_admin_account(context):
    response = context.supabase_client.sign_up("admin@admin.admin", "secret_password")
    admin_user = User(id="asd", email="admin@admin.admin", password="secret_password")
    admin = Player(
        user=admin_user,
        name="admin",
        jwt=response.session.access_token,
    )
    context.admin_http_client = RiskItClient(admin)


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
        player.connection.close()
    context.players = dict()
