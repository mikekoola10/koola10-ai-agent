import json
import os
from datetime import datetime

class EconomicLedger:
    def __init__(self, initial_balance: float = 100.0):
        self.storage_path = os.getenv("METACLAW_STORAGE_PATH", "/data")
        self.file_path = os.path.join(self.storage_path, "economic_ledger.json")
        self._ensure_storage_exists()
        self.data = self._load_data(initial_balance)

    def _ensure_storage_exists(self):
        if not os.path.exists(self.storage_path):
            try:
                os.makedirs(self.storage_path)
            except OSError:
                self.storage_path = "./data"
                self.file_path = os.path.join(self.storage_path, "economic_ledger.json")
                os.makedirs(self.storage_path, exist_ok=True)

    def _load_data(self, initial_balance):
        if os.path.exists(self.file_path):
            with open(self.file_path, 'r') as f:
                return json.load(f)
        return {
            "balance": initial_balance,
            "transactions": [],
            "costs": {}
        }

    def _save_data(self):
        with open(self.file_path, 'w') as f:
            json.dump(self.data, f, indent=2)

    def record_cost(self, category: str, amount: float, description: str):
        self.data["balance"] -= amount
        self.data["transactions"].append({
            "type": "COST",
            "category": category,
            "amount": amount,
            "description": description,
            "timestamp": datetime.now().isoformat()
        })
        self.data["costs"][category] = self.data["costs"].get(category, 0.0) + amount
        self._save_data()

    def record_revenue(self, amount: float, source: str):
        self.data["balance"] += amount
        self.data["transactions"].append({
            "type": "REVENUE",
            "amount": amount,
            "source": source,
            "timestamp": datetime.now().isoformat()
        })
        self._save_data()

    def get_balance(self) -> float:
        return self.data["balance"]

    def get_summary(self) -> dict:
        return {
            "balance": self.data["balance"],
            "total_expenses": sum(self.data["costs"].values()),
            "cost_breakdown": self.data["costs"],
            "transaction_count": len(self.data["transactions"])
        }
