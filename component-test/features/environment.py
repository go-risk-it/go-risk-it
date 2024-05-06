import logging
import time

from dotenv import load_dotenv
from src.client.rest.client import RiskItClient
from src.client.websockets.manager import RiskItWebsocketManager
from src.core.context import RiskItContext
from src.client.rest.prefix_session import PrefixSession
from src.core.runner import ServiceRunner

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
    context.websocket_manager = RiskItWebsocketManager()
    context.session = PrefixSession(prefix_url="http://localhost:8080")
    context.service_runner = ServiceRunner(
        start_command=start_command,
        path="../",
        timeout=10,
    )
    context.risk_it_client = RiskItClient(context.session)

    LOGGER.info("Starting service")
    context.service_runner.start()
    LOGGER.info("Service started")

    time.sleep(2)
