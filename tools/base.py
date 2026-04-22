from dataclasses import dataclass
from typing import Any, Optional

@dataclass
class ToolResult:
    success: bool
    output: Any
    error: Optional[str] = None
