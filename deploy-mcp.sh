#!/bin/bash
# Deploy MCP Hub to Fly.io

echo "🚀 Starting MCP Deployment..."

# 1. Verification
if [ ! -f "tools/mcp_client.go" ]; then
    echo "❌ tools/mcp_client.go missing!"
    exit 1
fi

# 2. Build Check
echo "🛠 Checking compilation..."
go build -o /dev/null main.go
if [ $? -ne 0 ]; then
    echo "❌ Compilation failed!"
    exit 1
fi

# 3. Deploy
echo "📦 Deploying to Fly.io..."
# flyctl deploy

echo "✅ MCP Deployment prepared."
