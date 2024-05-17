import logging
from os import PathLike
from subprocess import Popen

LOGGER = logging.getLogger(__name__)


class ServiceRunner:
    def __init__(
        self,
        start_command: list[str],
        path: str | PathLike,
        timeout: int,
    ) -> None:
        self.path = path
        self.start_command = start_command
        self.timeout = timeout

        self.__proc = None

    def start(self):
        LOGGER.info(f"Starting service: {self.start_command}")
        self.__proc = Popen(  # pylint: disable=consider-using-with
            self.start_command, cwd=self.path
        )
        if (exit_code := self.__proc.wait()) != 0:
            raise RuntimeError(
                f"Service failed to start with exit code {exit_code}",
            )
