from __future__ import annotations

from urllib.parse import urljoin

import requests


class PrefixSession(requests.Session):
    """Prefix every request url with the given ``prefix_url``.

    Recommended methods of using a fixed base URL, as specified in
    https://github.com/psf/requests/issues/2554#issuecomment-109341010.
    """

    def __init__(self, prefix_url: str) -> None:
        """Initialize BaseUrlSession."""
        super().__init__()
        self.prefix_url = prefix_url

    def request(self, method, url, *args, **kwargs) -> requests.Response:
        url = urljoin(self.prefix_url, url)
        return super().request(method, url, *args, **kwargs)
