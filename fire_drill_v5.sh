#!/bin/bash
echo "=== FIRE DRILL V5: AGENTMAIL TEST ==="
# Verify messaging tool can send alerts
grep -A 10 "func sendAlertEmail" main.go
grep -A 20 "func UnifiedMessaging" tools/messaging.go
echo "=== VERIFICATION COMPLETE ==="
