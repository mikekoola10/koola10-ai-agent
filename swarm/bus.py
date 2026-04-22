import redis
import json
import os

class SwarmBus:
    def __init__(self):
        self.redis_url = os.getenv("REDIS_URL", "redis://localhost:6379")
        try:
            self.client = redis.from_url(self.redis_url, decode_responses=True)
        except Exception:
            self.client = None

    def publish(self, channel: str, message: dict):
        if self.client:
            self.client.publish(channel, json.dumps(message))

    def subscribe(self, channel: str):
        if self.client:
            pubsub = self.client.pubsub()
            pubsub.subscribe(channel)
            return pubsub
        return None

    def get_recent_messages(self, channel: str, limit: int = 10):
        # Redis Pub/Sub is fire-and-forget, so for "recent messages"
        # in a minimal implementation we use a list
        if self.client:
            return [json.loads(m) for m in self.client.lrange(f"history:{channel}", -limit, -1)]
        return []

    def log_message(self, channel: str, message: dict):
        if self.client:
            self.client.rpush(f"history:{channel}", json.dumps(message))
            self.client.ltrim(f"history:{channel}", -100, -1)
