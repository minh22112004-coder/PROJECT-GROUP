import os
from functools import lru_cache

import redis
from django.conf import settings


@lru_cache(maxsize=1)
def get_redis_client():
    """
    Return a singleton Redis client configured from settings or environment.
    Defaults to localhost for non-Docker environments.
    Set REDIS_URL environment variable to override (e.g., redis://redis:6379/0 for Docker).
    """
    redis_url = getattr(
        settings,
        "REDIS_URL",
        os.environ.get("REDIS_URL", "redis://localhost:6379/0"),
    )
    return redis.Redis.from_url(redis_url)


