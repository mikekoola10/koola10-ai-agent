MAIN_PY_TEMPLATE = """from fastapi import FastAPI

app = FastAPI()

@app.get("/")
async def root():
    return {"message": "Welcome to {product_name}", "description": "{product_description}"}
"""

DOCKERFILE_TEMPLATE = """FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE 8080
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8080"]
"""

FLY_TOML_TEMPLATE = """app = "{product_name}-api"
primary_region = "ams"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 0
  processes = ["app"]

[[vm]]
  size = "shared-cpu-1x"
"""

REQUIREMENTS_TXT_TEMPLATE = """fastapi
uvicorn
"""

INDEX_HTML_TEMPLATE = """<!DOCTYPE html>
<html>
<head>
    <title>{product_name}</title>
</head>
<body>
    <h1>{product_name}</h1>
    <p>{product_description}</p>
</body>
</html>
"""
