# UIA-X Installation and Execution Script for Windows
# This script sets up the UIA-X MCP server on your local machine.

Write-Host "Starting UIA-X Installation..." -ForegroundColor Cyan

# 1. Clone the repository
if (-not (Test-Path "uia-x")) {
    Write-Host "Cloning UIA-X repository..."
    git clone https://github.com/doucej/uia-x.git
} else {
    Write-Host "UIA-X directory already exists."
}

cd uia-x

# 2. Create and activate virtual environment
if (-not (Test-Path ".venv")) {
    Write-Host "Creating virtual environment..."
    python -m venv .venv
}

Write-Host "Activating virtual environment and installing dependencies..."
& .\.venv\Scripts\Activate.ps1
pip install -e .

# 3. Set environment variables
$env:MCP_TRANSPORT="streamable-http"

# 4. Start the server
Write-Host "Starting UIA-X server on http://localhost:8000..." -ForegroundColor Green
python -m uiax.server
