# Grant Writing Skill
Full grant discovery → apply → submit workflow.

## Description
Automates the identification of grant opportunities, drafting of narratives based on organization profile, and submission of applications.

## Required Tools
- `grant` (Swarm)
- `ai_inference`
- `browser_automation`

## Trigger Conditions
- New grant matching keywords found in `grants.gov`.
- Manual trigger via `/services/grant/apply`.

## Step-by-Step Workflow
1. **Discovery**: `grant` swarm searches for relevant grants.
2. **Drafting**: `ai_inference` drafts application narrative using DeepSeek.
3. **Review**: Human or AI review of the draft.
4. **Submission**: `browser_automation` handles the form submission.

## Expected Output Format
`{ "application_id": "...", "status": "submitted", "grant_id": "..." }`
