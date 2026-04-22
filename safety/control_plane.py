import json
import os
from datetime import datetime

class RiskScorer:
    def __init__(self):
        self.risky_patterns = ["delete", "secrets", "production", "root", "sudo", "rm -rf"]

    def score(self, action: dict) -> float:
        score = 0.0
        # Check task description or payload for risky patterns
        action_str = str(action).lower()

        matches = [pattern for pattern in self.risky_patterns if pattern in action_str]

        if matches:
            # Each match adds 0.4 to the risk score
            score = min(1.0, len(matches) * 0.4)

        return score

class ControlPlane:
    def __init__(self, threshold: float = 0.6):
        self.threshold = threshold
        self.scorer = RiskScorer()
        self.storage_path = os.getenv("METACLAW_STORAGE_PATH", "/data")
        self.audit_log_path = os.path.join(self.storage_path, "control_plane_audit.jsonl")
        self._ensure_storage_exists()

    def _ensure_storage_exists(self):
        if not os.path.exists(self.storage_path):
            try:
                os.makedirs(self.storage_path)
            except OSError:
                self.storage_path = "./data"
                self.audit_log_path = os.path.join(self.storage_path, "control_plane_audit.jsonl")
                os.makedirs(self.storage_path, exist_ok=True)

    def evaluate(self, action: dict) -> dict:
        risk_score = self.scorer.score(action)
        decision = "ALLOW" if risk_score < self.threshold else "BLOCK"
        requires_review = risk_score >= 0.5

        result = {
            "timestamp": datetime.now().isoformat(),
            "action": action,
            "risk_score": risk_score,
            "decision": decision,
            "requires_review": requires_review
        }

        self._log_audit(result)
        return result

    def _log_audit(self, entry: dict):
        with open(self.audit_log_path, "a") as f:
            f.write(json.dumps(entry) + "\n")
