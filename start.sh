#!/bin/bash
set -e

# Start the Python browser agent in the background
echo "Starting Browser Agent on port 8081..."
export BROWSER_AGENT_PORT=8081
xvfb-run --server-args="-screen 0 1280x800x24" python3 browser-agent/main.py > /data/browser_agent.log 2>&1 &

# Wait for browser agent to start
sleep 5

# Start the Go backend
echo "Starting Go Backend on port 8080..."
./agent
