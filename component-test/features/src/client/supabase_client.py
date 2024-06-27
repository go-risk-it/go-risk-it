import logging
import os

from gotrue import AuthResponse
from gotrue.http_clients import SyncClient
from supabase import create_client, ClientOptions

from src.core.player import Player
from src.core.user import User

LOGGER = logging.getLogger(__name__)


class SupabaseClient:
    client: SyncClient

    def __init__(self):
        self.client = create_client(
            supabase_url=os.environ.get("SUPABASE_PUBLIC_URL"),
            supabase_key=os.environ.get("ANON_KEY"),
            options=ClientOptions(auto_refresh_token=False),
        )

    def get_user(self, email: str, password: str) -> User:
        LOGGER.info("Trying to sign in")
        response = self.__sign_in(email, password)

        if response is not None:
            LOGGER.info("User already registered")

            return User(id=response.user.id, email=response.user.email, jwt=response.session.access_token)

        LOGGER.info("User does not exist, signing up")
        response = self.__sign_up(email, password)

        return User(id=response.user.id, email=response.user.email, jwt=response.session.access_token)

    def __sign_up(self, email: str, password: str) -> AuthResponse:
        return self.client.auth.sign_up(
            credentials=dict(email=email, password=password)
        )

    def __sign_in(self, email: str, password: str) -> AuthResponse:
        return self.client.auth.sign_in_with_password(
            credentials=dict(email=email, password=password)
        )
