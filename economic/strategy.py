from typing import List
from economic.decision import EconomicDecisionEngine

class EconomicStrategy:
    def __init__(self, decision_engine: EconomicDecisionEngine):
        self.engine = decision_engine

    def allocate_budget(self, tasks: List[dict], total_budget: float) -> List[dict]:
        scored_tasks = []
        for task in tasks:
            eval_res = self.engine.should_execute(task)
            scored_tasks.append({
                "task": task,
                "roi": eval_res["roi"],
                "cost": eval_res["cost_estimate"]
            })

        # Sort by ROI descending
        scored_tasks.sort(key=lambda x: x["roi"], reverse=True)

        allocated = []
        remaining_budget = total_budget
        for item in scored_tasks:
            if item["cost"] <= remaining_budget:
                allocated.append(item["task"])
                remaining_budget -= item["cost"]

        return allocated

    def reinvest(self, balance: float, reinvestment_ratio: float = 0.3) -> float:
        return balance * reinvestment_ratio
