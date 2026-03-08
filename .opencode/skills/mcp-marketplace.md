# MCP Marketplace Skill

Use this skill when working on this repository. It defines the product intent, technical boundaries, and delivery priorities so implementation decisions stay aligned with the marketplace vision in `requirements.md`.

## Product Mission

Build an MCP marketplace that makes complex MCP servers easy for average users to discover, buy, install, run, and manage across clients like Claude Desktop, VS Code, Cursor, and Codex.

The platform must bridge two worlds:
- developer-grade MCP infrastructure and security
- consumer-friendly one-click onboarding and payments

## Core Personas

- Buyer: discovers servers, pays for access, installs with one click, manages hub connections, billing, and local/cloud usage
- Merchant: imports or builds MCP servers, configures auth/pricing/deployments, publishes listings, and monitors revenue and reliability
- Admin: governs tenants, audits security/payment activity, and manages platform-wide fee and payout policies

## What Success Looks Like

- Buyers can discover verified MCP servers and install them with minimal friction
- Merchants can publish both hosted and self-hostable MCP offerings
- OAuth, resource binding, and tenancy boundaries are enforced correctly
- Paid installs support real payment orchestration instead of dead-end entitlement failures
- The platform supports both cloud-hosted and local-agent installation paths

## Non-Negotiable Product Capabilities

### Marketplace
- public server discovery and detail pages
- verified listings, ratings, and filtering by local vs cloud support
- entitlement-aware install flow

### Buyer Experience
- one-click install using CIMD and DCR concepts
- personal MCP hub endpoint
- interactive connection flows for supported clients
- billing, invoices, wallet balance, top-ups, and payment preferences

### Merchant Experience
- Docker Hub image import or custom image publishing
- low-code agent builder via n8n integration
- deployment lifecycle: draft -> deployed -> published
- pricing, auth, observability, payouts, and revenue reporting

### Payments
- x402 per-call or install-time payment flow
- Stripe onramp for wallet top-ups
- Stripe Connect for seller onboarding and payouts
- platform-level fee policy support

### Security
- OAuth 2.1 + PKCE for remote installs
- resource/audience binding to canonical MCP server URIs
- multi-tenant isolation for users, tenants, servers, tokens, and billing records
- SSRF-aware metadata fetching and safe webhook validation

## Technical Direction

### Backend
- Primary backend is Go
- Go owns auth, API gateway behavior, tenancy checks, payments orchestration, MCP hub projection, and deployment APIs
- MongoDB is the intended persistent store
- In-memory store is only acceptable for tests or clearly insecure local development paths

### Frontend
- Primary frontend is Next.js with TypeScript
- Frontend should progressively replace mock data with live API integrations
- UX must favor clarity and guided actions over raw infrastructure exposure

### Orchestration
- n8n is the current low-code builder/deployment integration point
- Dockerized server deployment is part of the merchant lifecycle

## Delivery Priorities

When deciding what to work on next, prefer this order:

1. Broken core flows
   - paid install flow
   - auth/token install flow
   - server deployment/publish flow
2. Security correctness
   - JWT/OAuth metadata consistency
   - webhook verification
   - tenant and resource boundary enforcement
3. Product completeness for existing surfaces
   - wire UI to real backend APIs
   - remove placeholder payment states
   - improve buyer/merchant/admin operational views
4. Production hardening
   - persistence, observability, readiness checks, rate limiting, background work reliability

## Current Known Gaps

These are the most important known gaps derived from `requirements.md` and `docs/PAYMENTS_MISSING_FEATURES.md`:

- paid installs do not yet complete an end-to-end checkout and entitlement flow in the web app
- some advertised payment methods are not backed by complete provider adapters
- Stripe webhook verification needs stricter production enforcement
- buyer and merchant payment-method controls are incomplete in the UI
- x402 intent history and retry flows are not fully surfaced in the frontend
- some webapp pages still depend on mock or partial data wiring

## Implementation Rules For OpenCode

When making changes in this repository:

- preserve the buyer / merchant / admin role model
- preserve tenant isolation and do not introduce cross-tenant shortcuts
- prefer extending existing REST endpoints and store interfaces over parallel one-off patterns
- keep marketplace lifecycle rules intact: new listings start as draft, publish requires deployed + priced state
- treat payment method names as product-level contract values; avoid inventing new labels casually
- do not weaken OAuth, PKCE, resource binding, or payment verification for convenience
- prefer incremental, shippable work that closes real product gaps rather than broad speculative rewrites

## API And Data Expectations

- OAuth metadata must remain internally consistent across issuer, token endpoint, registration endpoint, protected resource metadata, and JWKS
- JWT/JWKS behavior must support client verification safely
- payment records, intents, payouts, and top-ups should be auditable
- mutating operations should continue to produce audit-friendly behavior

## Definition Of Done For Feature Work

A feature is only done when most of the following are true:

- backend logic exists and passes tests
- frontend path is wired to the real endpoint if a user-facing flow exists
- errors are actionable for buyers, merchants, or admins
- role checks and tenant checks are enforced
- docs or checklist notes are updated when the change affects product behavior

## Preferred Work Patterns

- For auth or payment changes, inspect both backend endpoints and frontend call sites before editing
- For UI work, connect to existing APIs before creating new placeholders
- For payment changes, verify method naming consistency across backend models, API payloads, and UI labels
- For install flow changes, validate both entitlement logic and client-facing install UX

## Quick Project Summary

This repo is not just a generic SaaS app. It is an MCP commerce and delivery platform. Favor decisions that improve:
- one-click installability
- secure cross-client authentication
- merchant deploy/publish workflows
- buyer payment and entitlement clarity
- production-grade multi-tenant behavior
