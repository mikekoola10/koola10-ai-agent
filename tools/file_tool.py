import os

def file_tool(payload):
    action = payload.get("action")
    path = payload.get("path")

    # Security: Restrict to /data or ./data and prevent path traversal
    storage_path = os.path.abspath(os.getenv("METACLAW_STORAGE_PATH", "/data"))
    if not os.path.exists(storage_path):
        os.makedirs(storage_path, exist_ok=True)

    abs_path = os.path.abspath(path)
    if not abs_path.startswith(storage_path):
        # Allow relative paths inside storage_path
        abs_path = os.path.abspath(os.path.join(storage_path, path))
        if not abs_path.startswith(storage_path):
            raise Exception(f"Security error: Path {path} is outside of allowed storage {storage_path}")

    if action == "read":
        if not os.path.exists(abs_path):
            raise Exception(f"File not found: {abs_path}")
        with open(abs_path, 'r') as f:
            return f.read()
    elif action == "write":
        content = payload.get("content", "")
        os.makedirs(os.path.dirname(abs_path), exist_ok=True)
        with open(abs_path, 'w') as f:
            f.write(content)
        return f"Successfully wrote to {abs_path}"
    else:
        raise Exception(f"Unsupported action: {action}")
