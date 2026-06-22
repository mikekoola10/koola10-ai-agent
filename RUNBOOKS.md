# Koola10 & Spiral Runbooks

This document outlines standard operating procedures for common failure modes in the Koola10 and Spiral ecosystems.

## 1. Financial Ledger Discrepancies
**Symptoms:** Balance doesn't match expected revenue; JSON parsing errors on startup; contamination between Koola10 and Spiral funds.

### Procedure: Recovery from Corruption
1. **Stop the Application**: `flyctl apps restart koola10` (if running).
2. **Inspect Data**: Use `cat /data/economic_ledger.json` and `cat /data/spiral_ledger.json`.
3. **Rollback**: If JSON is malformed, check for `.bak` files or restore from the last known good state in the audit log (`/data/audit_chain.jsonl`).
4. **Manual Reconcile**: Use the `reconcile_vault` tool if discrepancies exist between Stripe/Binance and the internal ledger.

## 2. AgentMail Downtime
**Symptoms:** Alerts not being received; BPA trial onboarding fails; `/admin/test_alert` returns `success: false`.

### Procedure: Manual Intervention
1. **Check Connectivity**: Verify `DEEPSEEK_API_KEY` and network access to `api.deepseek.com`.
2. **Fallback Logging**: All failed alerts are logged to `/data/audit_chain.jsonl`.
3. **Manual Onboarding**: For BPA trials, manually retrieve the `ApplicationID` from the logs and email the user their key.
4. **Restart**: Often, a restart resolves temporary tool timeouts.

## 3. Swarm Timeout / Machine Failure
**Symptoms:** `/monitor` shows high error rate or 0 active swarms; "all agents in vertical busy" errors.

### Procedure: Scalability & Recovery
1. **Restart Machine**: `flyctl apps restart koola10`. Note: The `Debugger Swarm` attempts this automatically if it detects cascading failures.
2. **Increase Capacity**: If "busy" errors persist, update `main.go` to deploy more agents in the specific vertical (e.g., `globalSwarmManager.DeploySwarms("leadgen", 20)`).
3. **Debugger Swarm Prerequisites**: For autonomous recovery via `flyctl`, the container must have `flyctl` installed and a valid `FLY_API_TOKEN` configured in the environment. If these are missing, the Debugger Swarm will log execution errors but cannot execute restarts.

## 4. Mirror Protocol Out of Sync
**Symptoms:** Agents making decisions against user preferences; `/mirror/context` returns stale data.

### Procedure: Preference Reset
1. **Clear Cache**: Remove `/data/user_mirror.json` to force a re-load from default "safe" settings.
2. **Trigger Reflection**: Use the `/ai/chat` endpoint to feed the agent new context, which will be saved to the Mirror profile.
