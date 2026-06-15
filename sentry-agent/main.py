import urllib.request
import json
import time
import os
import sys

TARGET_URL = os.getenv("TARGET_URL", "http://localhost:8080/health")
RECOVERY_URL = os.getenv("RECOVERY_URL", "http://localhost:8080/system/webhook/recovery")
SECRET = os.getenv("RECOVERY_WEBHOOK_SECRET")

print(f"--- SENTRY V2 STARTING (TARGET: {TARGET_URL}) ---")
sys.stdout.flush()

def check_health():
    try:
        with urllib.request.urlopen(TARGET_URL, timeout=5) as response:
            return response.getcode() == 200
    except Exception as e:
        print(f"[Sentry] Health check failed: {e}")
        sys.stdout.flush()
        return False

def trigger_recovery():
    print(f"[Sentry] CRITICAL FAILURE DETECTED. Triggering recovery...")
    sys.stdout.flush()
    data = json.dumps({
        "failure_name": "fly_app_down",
        "details": "Sentry detected health check failure",
        "secret": SECRET
    }).encode('utf-8')

    req = urllib.request.Request(RECOVERY_URL, data=data, headers={'Content-Type': 'application/json'}, method='POST')
    try:
        with urllib.request.urlopen(req, timeout=5) as response:
            print(f"[Sentry] Recovery webhook response: {response.getcode()}")
            sys.stdout.flush()
    except Exception as e:
        print(f"[Sentry] Failed to trigger recovery webhook: {e}")
        sys.stdout.flush()

def monitor():
    failures = 0
    while True:
        if check_health():
            print("[Sentry] Health check OK.")
            sys.stdout.flush()
            failures = 0
        else:
            failures += 1
            print(f"[Sentry] Failure count: {failures}")
            sys.stdout.flush()
            if failures >= 3:
                trigger_recovery()
                failures = 0

        time.sleep(5)

if __name__ == "__main__":
    if not SECRET:
        print("Error: RECOVERY_WEBHOOK_SECRET not set")
        sys.exit(1)
    monitor()
