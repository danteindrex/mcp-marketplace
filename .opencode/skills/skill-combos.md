# Skill Combos

Use this guide to load the right combination of project skills for common workflows.

## Core Rule

Always start with:
- `mcp-marketplace.md`

Then add the smallest set of domain skills needed for the task.

## Common Combos

### Paid Install Flow
- `buyer-install-journeys.md`
- `oauth-resource-security.md`
- `x402-payments.md`

Use when:
- a buyer must authenticate, pay, gain entitlement, and connect successfully

### OAuth Or Discovery Fixes
- `mcp-protocol-integration.md`
- `oauth-resource-security.md`

Use when:
- changing OAuth metadata, JWKS, PKCE, resource binding, or client compatibility behavior

### Merchant Deploy And Publish
- `n8n-agent-builder.md`
- `docker-marketplace-ops.md`
- `marketplace-trust-observability.md`

Use when:
- a seller imports an image, deploys a server, monitors failures, and publishes to the marketplace

### Python Server Scaffolding
- `fastmcp-server-factory.md`
- `mcp-protocol-integration.md`
- `docker-marketplace-ops.md`

Use when:
- generating or refining seller-facing MCP server templates and deployment-ready packages

### Buyer Onboarding And Client Compatibility
- `buyer-install-journeys.md`
- `mcp-protocol-integration.md`
- `oauth-resource-security.md`

Use when:
- improving one-click setup for Claude, Codex, VS Code, Cursor, or local bridge installs

### Payment Provider Readiness
- `x402-payments.md`
- `marketplace-trust-observability.md`

Use when:
- auditing method support, webhook readiness, settlement visibility, or payment ops gaps

### Marketplace Trust And Verification
- `marketplace-trust-observability.md`
- `buyer-install-journeys.md`

Use when:
- adding verification, ratings, listing trust signals, or reliability indicators

### Full Merchant Monetization Flow
- `docker-marketplace-ops.md`
- `n8n-agent-builder.md`
- `x402-payments.md`
- `marketplace-trust-observability.md`

Use when:
- taking a server from import -> deploy -> price -> publish -> payout readiness

### Production Hardening Pass
- `oauth-resource-security.md`
- `x402-payments.md`
- `marketplace-trust-observability.md`
- `docker-marketplace-ops.md`

Use when:
- closing security, telemetry, readiness, and operational reliability gaps before release

## Agent Pairings

Suggested matching agents for the combos above:
- paid install flow -> `buyer-experience-agent.md`, `oauth-security-auditor.md`, `payments-orchestrator.md`
- OAuth fixes -> `mcp-protocol-agent.md`, `oauth-security-auditor.md`
- merchant deploy/publish -> `merchant-lifecycle-agent.md`, `n8n-builder-agent.md`, `docker-ops-agent.md`
- Python server scaffolding -> `fastmcp-factory-agent.md`, `mcp-protocol-agent.md`
- trust and telemetry -> `trust-observability-agent.md`, `qa-release-agent.md`
