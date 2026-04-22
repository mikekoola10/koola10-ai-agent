import hashlib
import json
import os
from datetime import datetime

class ImmutableAuditLogger:
    def __init__(self):
        self.storage_path = os.getenv("METACLAW_STORAGE_PATH", "/data")
        self.log_file = os.path.join(self.storage_path, "immutable_audit.jsonl")
        self._ensure_storage_exists()

    def _ensure_storage_exists(self):
        if not os.path.exists(self.storage_path):
            os.makedirs(self.storage_path, exist_ok=True)

    def _get_last_hash(self):
        if not os.path.exists(self.log_file):
            return "genesis"
        with open(self.log_file, "rb") as f:
            f.seek(0, 2)
            if f.tell() == 0:
                return "genesis"
            # Read last line
            f.seek(-2, 2)
            while f.read(1) != b"\n":
                if f.tell() == 1:
                    f.seek(0)
                    break
                f.seek(-2, 1)
            last_line = f.readline().decode()
            if last_line:
                return json.loads(last_line).get("hash", "genesis")
        return "genesis"

    def log_event(self, event: dict):
        last_hash = self._get_last_hash()
        entry = {
            "timestamp": datetime.now().isoformat(),
            "event": event,
            "previous_hash": last_hash
        }
        entry_str = json.dumps(entry, sort_keys=True)
        entry["hash"] = hashlib.sha256(entry_str.encode()).hexdigest()

        with open(self.log_file, "a") as f:
            f.write(json.dumps(entry) + "\n")
        return entry["hash"]

    def verify_chain(self) -> bool:
        if not os.path.exists(self.log_file):
            return True

        expected_prev_hash = "genesis"
        with open(self.log_file, "r") as f:
            for line in f:
                entry = json.loads(line)
                actual_hash = entry.pop("hash")
                if entry["previous_hash"] != expected_prev_hash:
                    return False

                entry_str = json.dumps(entry, sort_keys=True)
                if hashlib.sha256(entry_str.encode()).hexdigest() != actual_hash:
                    return False

                expected_prev_hash = actual_hash
        return True

    def export_bundle(self):
        if not os.path.exists(self.log_file):
            return []
        with open(self.log_file, "r") as f:
            return [json.loads(line) for line in f]
