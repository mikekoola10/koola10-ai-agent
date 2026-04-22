import json
import time
from dataclasses import dataclass, asdict
from typing import List, Optional
import redis

@dataclass
class NodeInfo:
    node_id: str
    region: str
    endpoint: str
    last_seen: float

class CloudSwarmRegistry:
    def __init__(self, redis_client: redis.Redis):
        self.redis = redis_client
        self.key_prefix = "swarm:nodes:"
        self.ttl = 60 # Seconds

    def register(self, node_info: NodeInfo):
        key = f"{self.key_prefix}{node_info.node_id}"
        self.redis.set(key, json.dumps(asdict(node_info)), ex=self.ttl)

    def heartbeat(self, node_id: str):
        key = f"{self.key_prefix}{node_id}"
        data = self.redis.get(key)
        if data:
            info = json.loads(data)
            info["last_seen"] = time.time()
            self.redis.set(key, json.dumps(info), ex=self.ttl)

    def get_all_healthy_nodes(self) -> List[NodeInfo]:
        nodes = []
        keys = self.redis.keys(f"{self.key_prefix}*")
        for key in keys:
            data = self.redis.get(key)
            if data:
                nodes.append(NodeInfo(**json.loads(data)))
        return nodes

    def get_nodes_in_region(self, region: str) -> List[NodeInfo]:
        all_nodes = self.get_all_healthy_nodes()
        return [n for n in all_nodes if n.region == region]
