import json
import os
import uuid
from datetime import datetime

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
        self._save_data()
        return meeting_id

    def query_entity(self, name):
        return self.data["entities"].get(name)

    def get_all_meetings(self):
        return self.data["meetings"]
