import json
import os

class GlobalKillSwitch:
    def __init__(self, redis_client):
        self.redis = redis_client
        self.key = "compliance:killswitch"

    def trigger(self, reason: str, ttl: int = 3600):
        if self.redis:
            self.redis.set(self.key, reason, ex=ttl)

    def reset(self):
        if self.redis:
            self.redis.delete(self.key)

    def is_active(self) -> bool:
        if self.redis:
            return self.redis.exists(self.key) > 0
        return False

    def get_reason(self) -> str:
        if self.redis:
            val = self.redis.get(self.key)
            return val if val else "None"
        return "Redis unavailable"
