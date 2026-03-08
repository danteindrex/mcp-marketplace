# MCP Marketplace Agents

This directory defines project-specific OpenCode sub-agents for the MCP marketplace monorepo.

Use these agents to keep work specialized instead of mixing product, backend, frontend, payments, and security decisions in one prompt.

## Agents

- `product-strategist.md` - keeps work aligned with `requirements.md` and product priorities
- `backend-architect.md` - owns Go API, store, routing, and backend behavior
- `frontend-integrator.md` - owns Next.js UX and backend wiring
- `oauth-security-auditor.md` - reviews OAuth, JWT, JWKS, PKCE, and tenant/resource boundaries
- `payments-orchestrator.md` - owns x402, Stripe, settlement, payout, and payment UX flows
- `merchant-lifecycle-agent.md` - owns server import, deploy, publish, observability, and merchant workflows
- `buyer-experience-agent.md` - owns install flow, hub, connections, billing clarity, and client support
- `qa-release-agent.md` - owns test coverage, release readiness, regressions, and definition-of-done checks
- `mcp-protocol-agent.md` - owns MCP protocol behavior, transports, and cross-client compatibility
- `n8n-builder-agent.md` - owns low-code builder and workflow deployment sync
- `fastmcp-factory-agent.md` - owns Python/FastMCP server templates and generated scaffolds
- `docker-ops-agent.md` - owns Docker Hub import, image validation, and hosted/self-hosted readiness
- `trust-observability-agent.md` - owns verification signals, auditability, and operational telemetry

## Suggested Usage Pattern

- Start with `product-strategist.md` when planning work
- Use the specialist agent for implementation decisions in that domain
- Use the matching skill in `.opencode/skills/` alongside the specialist agent when the task is standards-heavy
- Finish with `qa-release-agent.md` for verification and gap checks

## Shared Context

All agents should assume:
- this repo is an MCP marketplace, not a generic CRUD SaaS app
- buyers, merchants, and admins have distinct workflows and permissions
- one-click install, secure auth, payments, and tenancy correctness matter more than cosmetic output
- `requirements.md`, `docs/API.md`, `docs/CHECKLIST.md`, and `docs/PAYMENTS_MISSING_FEATURES.md` are core project references
