import re

class DataBoundaryGuard:
    def __init__(self):
        self.pii_patterns = {
            "email": r"[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}",
            "ssn": r"\d{3}-\d{2}-\d{4}",
            "credit_card": r"\d{4}-\d{4}-\d{4}-\d{4}"
        }
        # In a real app, allowed regions would be in config
        self.allowed_regions = ["ams", "sjc", "fra", "local"]

    def check_region(self, region: str) -> bool:
        return region.lower() in self.allowed_regions

    def scan_for_pii(self, data: str) -> list:
        found = []
        for pii_type, pattern in self.pii_patterns.items():
            if re.search(pattern, data):
                found.append(pii_type)
        return found

    def redact_pii(self, data: str) -> str:
        redacted = data
        for pattern in self.pii_patterns.values():
            redacted = re.sub(pattern, "[REDACTED]", redacted)
        return redacted
