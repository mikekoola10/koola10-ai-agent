import re
import json
import os

def analyze_logs():
    log_file = "fly-logs.txt"
    if not os.path.exists(log_file):
        print(f"Log file {log_file} not found.")
        return

    with open(log_file, 'r', encoding='utf-8', errors='ignore') as f:
        logs = f.read()

    patterns = {
        "ModuleNotFoundError": r"ModuleNotFoundError: No module named '([\w\.\-]+)'",
        "ConnectionRefused": r"Connection refused",
        "OutOfMemory": r"Out of memory|OOM",
        "Panic": r"panic:",
        "InternalServerError": r"HTTP/1.1 500"
    }

    findings = []
    for error_type, pattern in patterns.items():
        matches = re.findall(pattern, logs)
        if matches:
            severity = "High" if error_type in ["OutOfMemory", "Panic", "ConnectionRefused"] else "Medium"
            findings.append({
                "type": error_type,
                "matches": list(set(matches)), # Use set to deduplicate
                "severity": severity
            })

    with open("analysis.json", "w") as f:
        json.dump(findings, f, indent=2)

    print(f"Found {len(findings)} issue types.")

if __name__ == "__main__":
    analyze_logs()
