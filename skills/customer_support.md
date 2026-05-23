# Customer Support Skill
Email/Slack → AI response → ticket resolution.

## Description
Automates the response and resolution of customer support inquiries.

## Required Tools
- `communication`
- `ai_inference`
- `developer` (for Jira/Issue integration)

## Trigger Conditions
- New incoming email to support address.
- Slack mention in `#support` channel.

## Step-by-Step Workflow
1. **Intake**: Detect new message via `communication` tool.
2. **Classification**: AI classifies the intent and urgency.
3. **Drafting**: AI drafts a response or technical solution.
4. **Action**: Respond via `communication` or create a ticket via `developer`.

## Expected Output Format
`{ "ticket_id": "SUP-42", "status": "resolved" }`
