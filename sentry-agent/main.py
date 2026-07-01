import time
import urllib.request
import os
import json

ORCHESTRATOR_URL = os.getenv("ORCHESTRATOR_URL", "http://koola10.fly.dev")
RECOVERY_WEBHOOK_SECRET = os.getenv("RECOVERY_WEBHOOK_SECRET")

def check_health():
    try:
        with urllib.request.urlopen(f"{ORCHESTRATOR_URL}/health", timeout=10) as response:
            if response.status == 200:
                print(f"Health check passed: {response.status}")
                return True
            else:
                print(f"Health check failed: {response.status}")
                return False
    except Exception as e:
        print(f"Health check error: {e}")
        return False

def trigger_recovery():
    print("Triggering system recovery...")
    url = f"{ORCHESTRATOR_URL}/system/webhook/recovery"
    headers = {
        "Content-Type": "application/json",
        "X-Recovery-Secret": RECOVERY_WEBHOOK_SECRET
    }
    data = json.dumps({"event": "CRITICAL_INFRASTRUCTURE_FAILURE", "mode": "fly_app_down"}).encode("utf-8")

    req = urllib.request.Request(url, data=data, headers=headers)
    try:
        with urllib.request.urlopen(req) as response:
            print(f"Recovery trigger response: {response.status}")
    except Exception as e:
        print(f"Failed to trigger recovery: {e}")

def main():
    print("Wizard's Sentry Agent starting...")
    while True:
        if not check_health():
            trigger_recovery()
        time.sleep(60) # Check every minute

if __name__ == "__main__":
    main()
