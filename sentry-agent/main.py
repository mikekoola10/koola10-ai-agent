import urllib.request
import json
import time
import os
import sys

# Independent Sentry - Fully Stateless
TARGET_URL = os.getenv("TARGET_URL", "https://koola10.fly.dev/health")
RECOVERY_URL = os.getenv("RECOVERY_URL", "https://koola10.fly.dev/system/webhook/recovery")
SECRET = os.getenv("RECOVERY_WEBHOOK_SECRET")

print(f"[Wizard's Sentry] Initialized. Monitoring: {TARGET_URL}")
sys.stdout.flush()

def check_health():
    try:
        with urllib.request.urlopen(TARGET_URL, timeout=10) as resp:
            return resp.getcode() == 200
    except Exception as e:
        print(f"[Sentry] Check failed: {e}")
        return False

def trigger_recovery():
    print(f"[Sentry] ALERT: Triggering remote recovery at {RECOVERY_URL}")
    data = json.dumps({
        "failure_name": "fly_app_down",
        "details": "Wizard's Sentry detected persistent unreachability",
        "secret": SECRET
    }).encode('utf-8')

    req = urllib.request.Request(RECOVERY_URL, data=data, headers={'Content-Type': 'application/json'}, method='POST')
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            print(f"[Sentry] Recovery Status: {resp.getcode()}")
    except Exception as e:
        print(f"[Sentry] Recovery Trigger Failed: {e}")
    sys.stdout.flush()

def monitor():
    failures = 0
    while True:
        if check_health():
            if failures > 0: print("[Sentry] Target reachable. Resetting failure counter.")
            failures = 0
        else:
            failures += 1
            print(f"[Sentry] Failure {failures}/3")
            if failures >= 3:
                trigger_recovery()
                failures = 0
        sys.stdout.flush()
        time.sleep(60)

if __name__ == "__main__":
    if not SECRET:
        print("[Sentry] FATAL: RECOVERY_WEBHOOK_SECRET missing.")
        sys.exit(1)
    monitor()
