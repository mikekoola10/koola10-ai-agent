import random
from typing import Optional
from swarm.cloud import CloudSwarmRegistry, NodeInfo

class TaskRouter:
    def __init__(self, registry: CloudSwarmRegistry):
        self.registry = registry

    def select_optimal_node(self, preferred_region: Optional[str] = None) -> Optional[NodeInfo]:
        nodes = []
        if preferred_region:
            nodes = self.registry.get_nodes_in_region(preferred_region)

        if not nodes:
            nodes = self.registry.get_all_healthy_nodes()

        if not nodes:
            return None

        # Select a random node from the available healthy nodes (could be improved with load balancing)
        return random.choice(nodes)
