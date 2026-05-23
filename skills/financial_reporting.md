# Financial Reporting Skill
Plaid + trading data → investor reports.

## Description
Aggregates financial data from banking APIs and internal trading swarms to produce comprehensive reports.

## Required Tools
- `finance`
- `trading` (Swarm)
- `financial_report` (Swarm)

## Trigger Conditions
- End of month/quarter.
- Investor request via dashboard.

## Step-by-Step Workflow
1. **Data Collection**: `finance` fetches latest bank balances and transactions.
2. **Trading P&L**: `trading` swarm reports daily profit/loss.
3. **Aggregation**: `financial_report` swarm consolidates data.
4. **Export**: Generate PDF or Markdown report for stakeholders.

## Expected Output Format
`{ "net_worth": 1250000.0, "monthly_profit": 15000.0, "status": "report_ready" }`
