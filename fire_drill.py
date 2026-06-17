import urllib.request
import json
import os
import time

def run_drill():
    print("--- FIRE DRILL START ---")
    timestamp = time.strftime('%Y-%m-%d %H:%M:%S', time.gmtime())
    print(f"Time: {timestamp} UTC")

    # 1. Health Check
    services = {
        "Koola10": "https://koola10.fly.dev/health",
        "Browser": "https://koola10-browser.fly.dev/health",
        "Semantic": "https://koola10-semantic.fly.dev/health",
        "Spiral": "https://spiral-ai-agent.onrender.com/"
    }

    results = {}
    for name, url in services.items():
        try:
            with urllib.request.urlopen(url, timeout=10) as response:
                status = response.getcode()
                results[name] = "ONLINE (200)" if status == 200 else f"ERROR ({status})"
        except Exception as e:
            results[name] = f"DOWN ({str(e)})"

    print("Health Status:", json.dumps(results, indent=2))

    # 2. Trigger Alert Email
    print("\nTriggering Alert Email via AgentMail Tool...")
    report_body = f"Fire Drill Report - {timestamp}\n\n"
    for name, status in results.items():
        report_body += f"{name}: {status}\n"

    payload = {
        "action": "send",
        "to": "mikekoola10@agentmail.to",
        "subject": f"FIRE DRILL: System Health Report {timestamp}",
        "body": report_body
    }

    try:
        req = urllib.request.Request(
            "https://koola10.fly.dev/tools/execute?tool_name=agentmail",
            data=json.dumps(payload).encode(),
            headers={"Content-Type": "application/json"},
            method='POST'
        )
        with urllib.request.urlopen(req) as response:
            res_data = json.loads(response.read().decode())
            print(f"Email Tool Response: {res_data}")
    except Exception as e:
        print(f"Failed to send drill email: {e}")

    print("--- FIRE DRILL COMPLETE ---")

if __name__ == "__main__":
    run_drill()
