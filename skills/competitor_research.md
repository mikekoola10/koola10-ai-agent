# Competitor Research Skill
Web scraping + analysis → strategic brief.

## Description
Monitors competitor activities and market trends to provide strategic insights.

## Required Tools
- `research` (Swarm)
- `browser_automation`

## Trigger Conditions
- Market volatility detected.
- Monthly competitor audit.

## Step-by-Step Workflow
1. **Scanning**: `browser_automation` visits competitor sites and news portals.
2. **Analysis**: `research` swarm extracts pricing changes and feature updates.
3. **Synthesis**: AI summarizes findings into a strategic brief.
4. **Notification**: Send brief to executives via `communication` (Slack).

## Expected Output Format
`{ "insights_found": 3, "status": "brief_sent" }`
