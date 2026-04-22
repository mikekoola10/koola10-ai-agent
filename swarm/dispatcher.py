from swarm.bus import SwarmBus

class SwarmDispatcher:
    def __init__(self, bus: SwarmBus):
        self.bus = bus

    def listen_and_dispatch(self):
        # In a production environment, this would run in a background thread or process
        # For this milestone, we provide the logic to pull from the bus
        pubsub = self.bus.subscribe("tasks")
        if pubsub:
            for message in pubsub.listen():
                if message['type'] == 'message':
                    # Simulated dispatch
                    print(f"Dispatcher received: {message['data']}")
