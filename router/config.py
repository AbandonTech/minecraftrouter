import os

import dotenv

dotenv.load_dotenv()

API_URL = os.getenv("API_URL")
API_TOKEN = os.getenv("API_TOKEN")

_missing_envs = [v[1] for v in ((API_URL, "API_URL"), (API_TOKEN, "API_TOKEN")) if v[0] is None]
if len(_missing_envs) > 0:
    raise EnvironmentError(f"Missing required environment variables: [{', '.join(repr(v) for v in _missing_envs)}]")

print("using API on", API_URL)
