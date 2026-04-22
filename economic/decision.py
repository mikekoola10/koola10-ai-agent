class EconomicDecisionEngine:
    def __init__(self):
        self.cost_map = {
            "product_generation": 0.50,
            "product_deployment": 0.10,
            "full_production": 0.60,
            "tool_execute": 0.02,
            "orchestration": 0.05,
            "semantic_search": 0.01,
            "reasoning": 0.03
        }

    def estimate_cost(self, action: dict) -> float:
        action_type = action.get("type", "unknown")
        return self.cost_map.get(action_type, 0.05)

    def estimate_value(self, action: dict) -> float:
        # Heuristic for potential revenue/value
        action_type = action.get("type", "unknown")
        if action_type in ["product_generation", "full_production"]:
            return 5.0 # High potential
        return 0.1 # Low individual utility

    def should_execute(self, action: dict, min_roi: float = 2.0) -> dict:
        cost = self.estimate_cost(action)
        value = self.estimate_value(action)

        roi = value / cost if cost > 0 else float('inf')

        recommendation = "ALLOW" if roi >= min_roi else "BLOCK"

        return {
            "recommendation": recommendation,
            "cost_estimate": cost,
            "value_estimate": value,
            "roi": roi,
            "min_roi_required": min_roi
        }
