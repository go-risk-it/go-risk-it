import logging
from abc import abstractmethod, ABC

import requests
from requests import RequestException

LOGGER = logging.getLogger(__name__)


class ReadinessCheck(ABC):
    """Base class for readiness checks."""

    @abstractmethod
    def is_ready(self) -> bool:
        """Abstract ``is_ready`` method.

        Concrete implementations of this method should implement the required logic
        to check if a service is ready and return a boolean indicating this status.
        """
        raise NotImplementedError()


class RestReadinessCheck(ReadinessCheck):
    """Readiness check that calls a REST endpoint."""

    def __init__(
            self,
            ready_endpoints: list[str],
            session: requests.Session | None = None,
    ) -> None:
        """Initialise ``RestReadinessCheck``.

        Args:
            ready_endpoints: List of endpoints that will return a 200 response
                when the service is ready.
            session: Session to use to call the ready
                endpoint, default is a bare ``requests.Session``.
        """
        self.__ready_endpoints = ready_endpoints
        self.__session = session if session else requests.Session()

    def is_ready(self) -> bool:
        """Execute healthcheck for ``RestReadinessCheck``."""
        try:
            for endpoint in self.__ready_endpoints:
                LOGGER.info("Performing health check by calling %s", endpoint)
                resp = self.__session.get(endpoint)
                LOGGER.info("Received health check response %s", resp.status_code)
                resp.raise_for_status()
        except RequestException as ex:
            LOGGER.info(ex)
            return False
        return True
