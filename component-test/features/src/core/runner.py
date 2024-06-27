import logging
from os import PathLike
from subprocess import Popen

from util.readiness import ReadinessCheck

LOGGER = logging.getLogger(__name__)


class ServiceRunner:
    def __init__(
            self,
            start_command: list[str],
            path: str | PathLike,
            timeout: int,
            readiness_check: ReadinessCheck | None = None,
    ) -> None:
        self.path = path
        self.start_command = start_command
        self.timeout = timeout
        self.__readiness_check = readiness_check

        self.__proc = None

    def start(self):
        print("Checking if service is already running")
        if self.__readiness_check:
            print ("readiness check configured, running it")

            if self.__readiness_check.is_ready():
                print("Service already running")
                return

        print("Service not running")

        print("Starting service")
        LOGGER.info(f"Starting service: {self.start_command}")
        self.__proc = Popen(  # pylint: disable=consider-using-with
            self.start_command, cwd=self.path
        )
        if (exit_code := self.__proc.wait()) != 0:
            raise RuntimeError(
                f"Service failed to start with exit code {exit_code}",
            )
