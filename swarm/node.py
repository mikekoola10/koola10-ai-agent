import os
from swarm.bus import SwarmBus

class SwarmNode:
    def __init__(self, bus: SwarmBus):
        self.bus = bus
        self.node_id = os.getenv("NODE_ID", "koola10-primary")

    def broadcast_task(self, task: str):
        msg = {"node_id": self.node_id, "task": task, "type": "task_broadcast"}
        self.bus.publish("tasks", msg)
        self.bus.log_message("tasks", msg)
        return msg

    def send_result(self, result: dict):
        msg = {"node_id": self.node_id, "result": result, "type": "task_result"}
        self.bus.publish("results", msg)
        self.bus.log_message("results", msg)
        return msg
