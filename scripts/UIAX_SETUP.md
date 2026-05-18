# UIA-X Installation and Execution Script for Windows

This script sets up the UIA-X MCP server on your local machine.

## Steps to set up your laptop:

1. **Prerequisites**: Ensure you have `git`, `python`, and `pip` installed.
2. **Download and Run**: Run the following PowerShell script on your laptop.

```powershell
# Create a dedicated directory
New-Item -ItemType Directory -Force -Path "$HOME\Koola10-Local"
Set-Location "$HOME\Koola10-Local"

Write-Host "Starting UIA-X Installation..." -ForegroundColor Cyan

# 1. Clone the repository
if (-not (Test-Path "uia-x")) {
    Write-Host "Cloning UIA-X repository..."
    git clone https://github.com/doucej/uia-x.git
}

cd uia-x

# 2. Create and activate virtual environment
if (-not (Test-Path ".venv")) {
    Write-Host "Creating virtual environment..."
    python -m venv .venv
}

Write-Host "Installing dependencies..."
& .\.venv\Scripts\Activate.ps1
pip install -e .

# 3. Configure for remote access
# To allow the Fly.io agent to reach your laptop, you'll need a tunnel like ngrok or cloudflared.
# Example with cloudflared: cloudflared tunnel --url http://localhost:8000
# Then set UIAX_SERVER_URL on Fly.io to the tunnel URL.

$env:MCP_TRANSPORT="streamable-http"

Write-Host "Starting UIA-X server on http://localhost:8000..." -ForegroundColor Green
Write-Host "NOTE: You MUST expose this port to the internet (e.g. via ngrok) for Koola10 on Fly.io to reach it." -ForegroundColor Yellow
Write-Host "Terminal automation is now enabled. Ensure your PowerShell execution policy allows running scripts." -ForegroundColor Cyan
python -m uiax.server
```

## Configuring Koola10 on Fly.io:

Once you have your local server running and exposed via a tunnel (e.g., `https://your-tunnel.ngrok-free.app`):

1. **Set the environment variables**:
   ```bash
   fly secrets set UIAX_SERVER_URL="https://your-tunnel.ngrok-free.app"
   fly secrets set UIAX_API_KEY="your-auto-generated-key"
   fly secrets set VERCEL_TOKEN="your-vercel-personal-access-token"
   ```

2. **Trigger the agent**:
   Send a POST request to `https://koola10.fly.dev/agent/watch-and-paste` with the window names:
   ```json
   {
     "source_window": "DeepSeek",
     "target_window": "Jules"
   }
   ```

## Terminal Automation Features:
Koola10 can now autonomously:
- Open a PowerShell terminal on your laptop.
- Execute `flyctl deploy` for backend updates.
- Execute `vercel --prod` for frontend updates.
- Detect "token" or "deploy" requests in your conversation and act without your intervention.
