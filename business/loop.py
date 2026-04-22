import json
import os
import uuid
from datetime import datetime

class BusinessLoop:
    def __init__(self):
        self.storage_path = os.getenv("METACLAW_STORAGE_PATH", "/data")
        self.file_path = os.path.join(self.storage_path, "business_loop.json")
        self._ensure_storage_exists()
        self.data = self._load_data()

    def _ensure_storage_exists(self):
        if not os.path.exists(self.storage_path):
            try:
                os.makedirs(self.storage_path)
            except OSError:
                self.storage_path = "./data"
                self.file_path = os.path.join(self.storage_path, "business_loop.json")
                os.makedirs(self.storage_path, exist_ok=True)

    def _load_data(self):
        if os.path.exists(self.file_path):
            with open(self.file_path, 'r') as f:
                return json.load(f)
        return {"actions": {}}

    def _save_data(self):
        with open(self.file_path, 'w') as f:
            json.dump(self.data, f, indent=2)

    def log_action(self, action, result):
        action_id = str(uuid.uuid4())
        self.data["actions"][action_id] = {
            "action": action,
            "result": result,
            "outcome": None,
            "timestamp": datetime.now().isoformat()
        }
        self._save_data()
        return action_id

    def log_outcome(self, action_id, outcome):
        if action_id in self.data["actions"]:
            self.data["actions"][action_id]["outcome"] = outcome
            self._save_data()
            return True
        return False

    def get_feedback_signal(self) -> dict:
        actions = list(self.data["actions"].values())
        if not actions:
            return {"success_rate": 0.0, "total_actions": 0, "recent_actions": []}

        outcomes = [a["outcome"] for a in actions if a["outcome"] is not None]
        successes = [o for o in outcomes if o.get("success") is True]

        success_rate = len(successes) / len(outcomes) if outcomes else 0.0

        return {
            "success_rate": success_rate,
            "total_actions": len(actions),
            "outcomes_reported": len(outcomes),
            "recent_actions": actions[-10:]
        }
