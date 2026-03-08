# OAuth Resource Security Skill

Use this skill when touching OAuth, PKCE, JWT, JWKS, discovery metadata, resource binding, or multi-tenant auth.

## Why This Skill Exists

Research notes:
- RFC 7636 requires PKCE for public clients and strongly prefers `S256`.
- RFC 8707 defines the `resource` parameter and recommends audience-restricting tokens to the requested resource.
- RFC 8414 defines authorization server metadata like `issuer`, `token_endpoint`, `registration_endpoint`, `jwks_uri`, and `code_challenge_methods_supported`.
- MCP remote HTTP transport documentation recommends standard HTTP authentication, with OAuth being the recommended path.

## What This Repo Needs

- standards-consistent OAuth discovery and token issuance
- public-client-safe install flows for VS Code, Cursor, Codex, and similar tools
- strict token audience/resource binding to canonical MCP server URIs
- tenant-aware authorization that prevents confused-deputy and cross-tenant token misuse

## Decision Rules

- use PKCE with `S256` by default for public clients
- keep `issuer`, metadata endpoints, and `jwks_uri` internally consistent
- bind access tokens to the canonical protected resource URI
- prefer one audience per token unless there is a compelling trust reason otherwise
- never weaken auth validation for convenience in production flows
- require exact subject, tenant, and resource boundary checks where applicable

## Implementation Focus

- `/.well-known/oauth-authorization-server`
- `/.well-known/oauth-protected-resource`
- `/.well-known/jwks.json`
- authorization code flow and DCR/CIMD-related install behavior
- token claims, `aud`, `kid`, and metadata correctness
- WWW-Authenticate and insufficient scope handling

## Done Looks Like

- clients can discover auth metadata reliably
- tokens are resource-bound and verifiable
- PKCE is enforced correctly
- cross-client identity works without breaking tenant isolation
- errors like `invalid_target` and `insufficient_scope` are actionable
