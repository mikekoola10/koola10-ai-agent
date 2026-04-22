import subprocess

def github_tool(payload):
    action = payload.get("action")
    if action == "list_repos":
        result = subprocess.run(["gh", "repo", "list"], capture_output=True, text=True)
        if result.returncode != 0:
            raise Exception(result.stderr)
        return result.stdout
    else:
        raise Exception(f"Unsupported action: {action}")
