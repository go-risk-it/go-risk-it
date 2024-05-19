import os

from gotrue import (
    SignUpWithEmailAndPasswordCredentials,
    SignInWithEmailAndPasswordCredentials,
)
from supabase import create_client
from supabase._sync.client import SyncClient


class SupabaseClient:
    client: SyncClient

    def __init__(self):
        self.client = create_client(
            supabase_url=os.environ.get("SUPABASE_PUBLIC_URL"),
            supabase_key=os.environ.get("ANON_KEY"),
        )

    def sign_up(self, email: str, password: str):
        return self.client.auth.sign_up(
            credentials=SignUpWithEmailAndPasswordCredentials(
                email=email, password=password
            )
        )

    def sign_in(self, email: str, password: str):
        return self.client.auth.sign_in_with_password(
            credentials=SignInWithEmailAndPasswordCredentials(
                email=email, password=password
            )
        )

    def close(self):
        self.client.auth.close()
