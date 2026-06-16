#!/bin/bash
set -e

echo "Starting Koola10 Smoke Test..."

# Check Health
HEALTH=$(curl -s http://localhost:8080/health)
if [[ "$HEALTH" == *"ok"* ]]; then
  echo "✅ Health check passed"
else
  echo "❌ Health check failed"
  exit 1
fi

# Check AGI Knowledge
AGI=$(curl -s http://localhost:8080/agi/knowledge)
if [[ "$AGI" == *"["* ]]; then
  echo "✅ AGI Knowledge endpoint is active"
else
  echo "❌ AGI Knowledge endpoint failed"
  exit 1
fi

# Check Ledger Summary
LEDGER=$(curl -s http://localhost:8080/economic/ledger/summary)
if [[ "$LEDGER" == *"balance"* ]]; then
  echo "✅ Economic Ledger endpoint is active"
else
  echo "❌ Economic Ledger endpoint failed"
  exit 1
fi

echo "Smoke Test Completed Successfully!"
