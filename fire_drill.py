import urllib.request
import json
import os
import time

API_KEY = os.getenv("AGENTMAIL_API_KEY")
INBOX_ID = os.getenv("AGENTMAIL_INBOX_ID", "mikekoola10@agentmail.to")
URL = f"https://api.agentmail.to/v0/inboxes/{INBOX_ID}/messages/send"

def send_alert(message):
    print(f"Sending alert: {message}")
    payload = {
        "to": "mikekoola10@agentmail.to",
        "subject": "Fire Drill Alert",
        "text": message
    }
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    try:
        req = urllib.request.Request(URL, data=json.dumps(payload).encode(), headers=headers, method='POST')
        with urllib.request.urlopen(req) as response:
            print(f"Alert sent. Status: {response.getcode()}")
    except Exception as e:
        print(f"Failed to send alert: {e}")

def run_drill():
    print("--- FIRE DRILL START ---")
    send_alert("Simulated failure: Redis connection lost. Engine initiating recovery.")
    print("--- FIRE DRILL COMPLETE ---")

if __name__ == "__main__":
    run_drill()
