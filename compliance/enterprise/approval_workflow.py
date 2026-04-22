import uuid
from datetime import datetime

class ApprovalWorkflow:
    def __init__(self):
        self.requests = {}

    def create_request(self, action: dict, approvers: list) -> str:
        request_id = str(uuid.uuid4())
        self.requests[request_id] = {
            "id": request_id,
            "action": action,
            "approvers": {a: False for a in approvers},
            "status": "PENDING",
            "timestamp": datetime.now().isoformat()
        }
        return request_id

    def approve(self, request_id: str, approver: str) -> bool:
        if request_id in self.requests and approver in self.requests[request_id]["approvers"]:
            self.requests[request_id]["approvers"][approver] = True
            # If all approved
            if all(self.requests[request_id]["approvers"].values()):
                self.requests[request_id]["status"] = "APPROVED"
            return True
        return False

    def check_status(self, request_id: str) -> dict:
        return self.requests.get(request_id, {"status": "NOT_FOUND"})
