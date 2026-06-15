#!/bin/bash
# Koola10 Smoke Test Script
set -e

KOOLA_URL=${KOOLA_URL:-"https://koola10.fly.dev"}
SPIRAL_URL=${SPIRAL_URL:-"https://spiral-ai-agent.onrender.com"}

echo "=== Starting Smoke Test Verification ==="

# 1. Koola10 Orchestrator Health
echo -n "Checking Koola10 Health... "
K_HEALTH=$(curl -s -o /dev/null -w "%{http_code}" "$KOOLA_URL/health")
if [ "$K_HEALTH" == "200" ]; then
    echo "[OK]"
else
    echo "[FAILED: $K_HEALTH]"
    exit 1
fi

# 2. Koola10 AI Chat Check
echo -n "Checking Koola10 AI Chat... "
K_CHAT=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$KOOLA_URL/ai/chat" -H "Content-Type: application/json" -d '{"prompt": "smoke test"}')
if [ "$K_CHAT" == "200" ]; then
    echo "[OK]"
else
    echo "[FAILED: $K_CHAT]"
    exit 1
fi

# 3. Spiral Browser Agent Health
echo -n "Checking Spiral Health... "
S_HEALTH=$(curl -s -o /dev/null -w "%{http_code}" "$SPIRAL_URL/")
if [ "$S_HEALTH" == "200" ]; then
    echo "[OK]"
else
    echo "[FAILED: $S_HEALTH]"
    exit 1
fi

echo "=== Smoke Test PASSED ==="
