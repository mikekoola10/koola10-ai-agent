from typing import Callable, Dict, Any
from tools.base import ToolResult

class ToolRegistry:
    def __init__(self):
        self.tools: Dict[str, Callable] = {}

    def register(self, name: str, func: Callable):
        self.tools[name] = func

    def run(self, name: str, payload: Any) -> ToolResult:
        if name not in self.tools:
            return ToolResult(success=False, output=None, error=f"Tool '{name}' not found")

        try:
            output = self.tools[name](payload)
            return ToolResult(success=True, output=output)
        except Exception as e:
            return ToolResult(success=False, output=None, error=str(e))
