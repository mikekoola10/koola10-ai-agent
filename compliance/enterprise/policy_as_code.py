import yaml
import os

class PolicyAsCode:
    def __init__(self):
        self.storage_path = os.getenv("METACLAW_STORAGE_PATH", "/data")
        self.policies_path = os.path.join(self.storage_path, "policies")
        os.makedirs(self.policies_path, exist_ok=True)
        self._create_default_policy()

    def _create_default_policy(self):
        default_policy = {
            "name": "default-security",
            "rules": [
                {"action": "delete", "effect": "BLOCK"},
                {"action": "rm -rf", "effect": "BLOCK"},
                {"action": "secrets", "effect": "BLOCK"}
            ]
        }
        policy_file = os.path.join(self.policies_path, "default.yaml")
        if not os.path.exists(policy_file):
            with open(policy_file, "w") as f:
                yaml.dump(default_policy, f)

    def evaluate(self, action: dict) -> dict:
        action_str = str(action).lower()

        for filename in os.listdir(self.policies_path):
            if filename.endswith(".yaml"):
                with open(os.path.join(self.policies_path, filename), "r") as f:
                    policy = yaml.safe_load(f)
                    for rule in policy.get("rules", []):
                        if rule["action"] in action_str:
                            return {"decision": rule["effect"], "policy": policy["name"], "rule": rule["action"]}

        return {"decision": "ALLOW"}

    def add_policy(self, name: str, rules: list):
        policy_file = os.path.join(self.policies_path, f"{name}.yaml")
        with open(policy_file, "w") as f:
            yaml.dump({"name": name, "rules": rules}, f)
