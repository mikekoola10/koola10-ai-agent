# Compliance Monitoring Skill
SOC2/GDPR/HIPAA continuous checks.

## Description
Continuously monitors system state and logs to ensure compliance with major standards.

## Required Tools
- `compliance` (Swarm)
- `audit_logger`

## Trigger Conditions
- System configuration change.
- Recurring hourly check.

## Step-by-Step Workflow
1. **Audit**: `compliance` swarm reviews recent `audit_chain.jsonl` entries.
2. **Scan**: Verifies encryption and access control settings.
3. **Alert**: Log any deviations or potential risks.
4. **Report**: Generate compliance status report.

## Expected Output Format
`{ "compliance_score": 98, "status": "all_checks_passed" }`
