import os

from gotrue import AuthResponse
from gotrue.http_clients import SyncClient
from supabase import create_client, ClientOptions


class SupabaseClient:
    client: SyncClient

    def __init__(self):
        self.client = create_client(
            supabase_url=os.environ.get("SUPABASE_PUBLIC_URL"),
            supabase_key=os.environ.get("ANON_KEY"),
            options=ClientOptions(auto_refresh_token=False),
        )

    def sign_up(self, email: str, password: str) -> AuthResponse:
        return self.client.auth.sign_up(
            credentials=dict(email=email, password=password)
        )

    def sign_in(self, email: str, password: str):
        return self.client.auth.sign_in_with_password(
            credentials=dict(email=email, password=password)
        )
