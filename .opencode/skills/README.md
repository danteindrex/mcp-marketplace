# OpenCode Skills

These skills give OpenCode project-specific operating context for the MCP marketplace.

## Core Skills

- `mcp-marketplace.md` - overall product brief and delivery priorities
- `mcp-protocol-integration.md` - MCP architecture, transports, primitives, and client compatibility
- `oauth-resource-security.md` - OAuth metadata, PKCE, JWKS, resource indicators, and tenant-safe token design
- `x402-payments.md` - HTTP 402 payment flows, settlement, wallets, and paid install behavior
- `n8n-agent-builder.md` - low-code builder and deployment workflow design around n8n
- `fastmcp-server-factory.md` - Python/FastMCP server creation and hosted-server expectations
- `docker-marketplace-ops.md` - Docker Hub import, image validation, packaging, and hosting rules
- `buyer-install-journeys.md` - one-click install, client UX, and post-install recovery flows
- `marketplace-trust-observability.md` - verification, ratings, auditability, and operational telemetry
- `skill-combos.md` - recommended skill bundles and matching agent pairings for common workflows

## Suggested Usage

- Start with `mcp-marketplace.md`
- Load one or more domain skills for the task at hand
- When work spans install + auth + payments, combine:
  - `buyer-install-journeys.md`
  - `oauth-resource-security.md`
  - `x402-payments.md`
