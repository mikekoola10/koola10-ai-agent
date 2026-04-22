import os
import subprocess
import re

from production.templates import (
    MAIN_PY_TEMPLATE, DOCKERFILE_TEMPLATE, FLY_TOML_TEMPLATE,
    REQUIREMENTS_TXT_TEMPLATE, INDEX_HTML_TEMPLATE
)

class ProductGenerator:
    def __init__(self):
        self.storage_path = os.path.abspath(os.getenv("METACLAW_STORAGE_PATH", "/data"))
        self.products_path = os.path.join(self.storage_path, "products")

    def generate(self, spec: dict) -> dict:
        raw_name = spec.get("name", "unknown-product")
        # Security: Sanitize product name to prevent path traversal
        product_name = re.sub(r'[^a-zA-Z0-9\-]', '', raw_name.lower().replace(" ", "-"))
        if not product_name:
            product_name = "unknown-product"

        product_description = spec.get("description", "An autonomous product")

        product_dir = os.path.join(self.products_path, product_name)
        os.makedirs(product_dir, exist_ok=True)

        # Robustness: Manual replacement instead of .format() to avoid crashes on braces in descriptions
        files = {
            "main.py": MAIN_PY_TEMPLATE.replace("{product_name}", product_name).replace("{product_description}", product_description),
            "Dockerfile": DOCKERFILE_TEMPLATE,
            "fly.toml": FLY_TOML_TEMPLATE.replace("{product_name}", product_name),
            "requirements.txt": REQUIREMENTS_TXT_TEMPLATE,
            "index.html": INDEX_HTML_TEMPLATE.replace("{product_name}", product_name).replace("{product_description}", product_description)
        }

        written_files = []
        for filename, content in files.items():
            filepath = os.path.join(product_dir, filename)
            with open(filepath, "w") as f:
                f.write(content)
            written_files.append(filename)

        # Initialize git repo
        try:
            subprocess.run(["git", "init"], cwd=product_dir, capture_output=True)
            subprocess.run(["git", "add", "."], cwd=product_dir, capture_output=True)
            subprocess.run(["git", "commit", "-m", "Initial product generation"], cwd=product_dir, capture_output=True)
        except Exception as e:
            print(f"Git initialization failed: {e}")

        return {
            "product_name": product_name,
            "directory": product_dir,
            "files": written_files
        }
