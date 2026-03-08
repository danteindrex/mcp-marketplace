# n8n Agent Builder Skill

Use this skill when working on the low-code builder, deployment automation, workflow import/export, or merchant authoring UX around n8n.

## Why This Skill Exists

Research notes:
- n8n is a workflow automation platform with strong support for visual workflow editing, webhooks, RBAC, sharing, source control, and AI-oriented nodes.
- n8n docs emphasize workflows, credentials, user management, project sharing, log streaming, and MCP-related nodes.

## What This Repo Needs

- a merchant-friendly builder that turns workflow creation into a marketplace-ready product path
- reliable deployment synchronization between n8n workflows and published marketplace servers
- clear status and failure visibility for deployment jobs

## Decision Rules

- keep builder UX focused on merchant outcomes, not raw n8n internals
- treat credentials, secrets, and publishing controls as separate concerns from workflow editing
- keep deploy status explicit: queued, retrying, failed, deployed
- preserve draft and publish gates even if a workflow exists in n8n

## Implementation Focus

- n8n workflow creation and activation integration
- deployment queue state and retries
- workflow identity storage and sync behavior
- merchant-facing builder entry points and deployment status surfaces
- future support for templated agent/server generation

## Done Looks Like

- merchants can create or link workflows without unclear handoffs
- deployment status is visible and trustworthy
- workflow activation does not bypass marketplace lifecycle rules
- failures provide concrete next steps
