from behave.runner import Context

from util.client import RiskItClient
from util.prefix_session import PrefixSession
from util.runner import ServiceRunner


class RiskItContext(Context):
    session: PrefixSession
    risk_it_client: RiskItClient
    service_runner: ServiceRunner
