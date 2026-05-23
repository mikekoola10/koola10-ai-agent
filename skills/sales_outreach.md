# Sales Outreach Skill
Lead scoring → personalized emails → follow-up automation.

## Description
Drives sales by scoring leads and executing personalized outreach sequences.

## Required Tools
- `crm`
- `ai_inference`
- `communication`

## Trigger Conditions
- Lead score exceeds threshold in CRM.
- Manual list import.

## Step-by-Step Workflow
1. **Scoring**: `crm` tool retrieves lead data for AI scoring.
2. **Personalization**: AI drafts personalized email based on lead's background.
3. **Outreach**: `communication` sends the initial outreach.
4. **Follow-up**: Automate follow-up if no response within 3 days.

## Expected Output Format
`{ "sequence_id": "outreach_99", "leads_contacted": 25 }`
