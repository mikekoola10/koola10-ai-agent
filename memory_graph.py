import json
import os
import uuid
from datetime import datetime
from typing import List, Dict, Optional, Any

class MemoryGraph:
    def __init__(self):
        self.storage_path = os.getenv("METACLAW_STORAGE_PATH", "/data")
        self.file_path = os.path.join(self.storage_path, "memory_graph.json")
        self._ensure_storage_exists()
        self.data = self._load_data()

    def _ensure_storage_exists(self):
        if not os.path.exists(self.storage_path):
            try:
                os.makedirs(self.storage_path)
            except OSError:
                # Fallback to local directory if /data is not writable
                self.storage_path = "./data"
                self.file_path = os.path.join(self.storage_path, "memory_graph.json")
                os.makedirs(self.storage_path, exist_ok=True)

    def _load_data(self):
        if os.path.exists(self.file_path):
            with open(self.file_path, 'r') as f:
                return json.load(f)
        return {"meetings": [], "entities": {}, "edges": []}

    def _save_data(self):
        with open(self.file_path, 'w') as f:
            json.dump(self.data, f, indent=2)

    def add_meeting(self, transcript):
        meeting_id = str(uuid.uuid4())
        meeting = {
            "id": meeting_id,
            "transcript": transcript,
            "timestamp": datetime.now().isoformat()
        }
        self.data["meetings"].append(meeting)
        # For Milestone 5, let's auto-generate some entities and edges for testing if they don't exist
        # In a real app, these would come from the LLM reasoning layer
        self._save_data()
        return meeting_id

    def query_entity(self, name):
        return self.data["entities"].get(name)

    def get_all_meetings(self):
        return self.data["meetings"]

    def calculate_influence_score(self, entity: str) -> float:
        score = 0.0
        for edge in self.data.get("edges", []):
            if edge["source"] == entity:
                score += edge.get("weight", 1.0)
            if edge["target"] == entity:
                score += edge.get("weight", 1.0)
        return score

    def find_causal_chains(self, start_entity: str, max_depth: int = 4) -> List[Dict]:
        chains = []
        queue = [(start_entity, [], 0.0)] # (current_entity, path, cumulative_weight)

        while queue:
            curr, path, weight = queue.pop(0)
            if len(path) >= max_depth:
                continue

            for edge in self.data.get("edges", []):
                if edge["source"] == curr:
                    new_path = path + [edge]
                    new_weight = weight + edge.get("weight", 1.0)
                    chains.append({
                        "path": new_path,
                        "cumulative_weight": new_weight
                    })
                    queue.append((edge["target"], new_path, new_weight))
        return chains

    def rank_decisions_by_impact(self) -> List[Dict]:
        decisions = []
        for entity_name, entity_data in self.data.get("entities", {}).items():
            if entity_data.get("type") == "decision":
                score = self.calculate_influence_score(entity_name)
                decisions.append({
                    "entity": entity_name,
                    "score": score,
                    "data": entity_data
                })

        decisions.sort(key=lambda x: x["score"], reverse=True)
        return decisions

    def find_path(self, source: str, target: str, max_depth: int = 3) -> Optional[List[Dict]]:
        queue = [(source, [])]
        visited = set()

        while queue:
            curr, path = queue.pop(0)
            if curr == target:
                return path

            if len(path) >= max_depth:
                continue

            if curr not in visited:
                visited.add(curr)
                for edge in self.data.get("edges", []):
                    if edge["source"] == curr:
                        queue.append((edge["target"], path + [edge]))
        return None
