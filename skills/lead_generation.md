# Lead Generation Skill
CRM scraping → enrichment → outreach sequence.

## Description
Finds potential leads, enriches them with additional data, and initiates outreach campaigns.

## Required Tools
- `leadgen` (Swarm)
- `crm`
- `communication`

## Trigger Conditions
- New target industry identified.
- Daily scheduled task at 9 AM.

## Step-by-Step Workflow
1. **Scraping**: `leadgen` swarm searches for prospects.
2. **Enrichment**: Additional details (email, LinkedIn) are fetched.
3. **CRM Sync**: Leads are created in `crm` (Salesforce/HubSpot).
4. **Outreach**: `communication` sends personalized emails or Slack messages.

## Expected Output Format
`{ "leads_generated": 50, "status": "outreach_initiated" }`
