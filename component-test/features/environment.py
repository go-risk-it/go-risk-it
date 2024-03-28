import logging

from util.client import RiskItClient
from util.context import RiskItContext
from util.prefix_session import PrefixSession
from util.runner import ServiceRunner

LOGGER = logging.getLogger(__name__)


def before_all(context: RiskItContext):
    start_command = [
        "docker-compose",
        "--project-name",
        "component-test",
        "--file",
        "docker-compose.yml",
        "up",
        "--build",
        "-d",
    ]

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
