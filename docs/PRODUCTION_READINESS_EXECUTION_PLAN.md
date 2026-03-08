# Production Readiness Execution Plan

This document consolidates implementation plans drafted by the project-specific OpenCode agents for bringing this MCP marketplace to full production readiness.

Contributing planning agents:
- `product-strategist`
- `oauth-security-auditor`
- `payments-orchestrator`
- `mcp-protocol-agent`
- `backend-architect`
- `frontend-integrator`
- `merchant-lifecycle-agent`
- `buyer-experience-agent`
- `trust-observability-agent`
- `qa-release-agent`

## Program Rules

- No mock production paths
- No placeholder provider support in buyer, merchant, or admin surfaces
- No lifecycle bypasses for publishability
- No browser-managed bearer-token auth model
- No trust badge, compatibility claim, or payment method shown without backend truth behind it

## Skill And Agent Routing

### Milestone 1: Auth And Security Baseline
- Skills: `mcp-marketplace.md`, `oauth-resource-security.md`, `mcp-protocol-integration.md`
- Agents: `oauth-security-auditor`, `backend-architect`, `frontend-integrator`

### Milestone 2: OAuth Providers And Client Install Interoperability
- Skills: `mcp-marketplace.md`, `oauth-resource-security.md`, `buyer-install-journeys.md`, `mcp-protocol-integration.md`
- Agents: `oauth-security-auditor`, `buyer-experience-agent`, `mcp-protocol-agent`, `frontend-integrator`

### Milestone 3: Paid Install And Payment Completion
- Skills: `mcp-marketplace.md`, `x402-payments.md`, `buyer-install-journeys.md`, `marketplace-trust-observability.md`
- Agents: `payments-orchestrator`, `buyer-experience-agent`, `backend-architect`, `frontend-integrator`, `trust-observability-agent`

### Milestone 4: Buyer Readiness
- Skills: `mcp-marketplace.md`, `buyer-install-journeys.md`, `mcp-protocol-integration.md`, `oauth-resource-security.md`
- Agents: `buyer-experience-agent`, `frontend-integrator`, `mcp-protocol-agent`

### Milestone 5: Merchant Readiness
- Skills: `mcp-marketplace.md`, `docker-marketplace-ops.md`, `n8n-agent-builder.md`, `x402-payments.md`, `marketplace-trust-observability.md`
- Agents: `merchant-lifecycle-agent`, `docker-ops-agent`, `n8n-builder-agent`, `payments-orchestrator`, `frontend-integrator`

### Milestone 6: Trust, Operations, And Release Hardening
- Skills: `mcp-marketplace.md`, `marketplace-trust-observability.md`, `oauth-resource-security.md`, `docker-marketplace-ops.md`
- Agents: `trust-observability-agent`, `backend-architect`, `qa-release-agent`

## Milestone 1: Auth And Security Baseline

### Outcome
- Production-safe identity, session, OAuth, and JWT/JWKS behavior that all later install and payment flows can trust.

### Implementation Scope
- Replace the current single bearer-token model with short-lived access tokens plus durable session/refresh state.
- Split browser session tokens from OAuth MCP resource tokens.
- Require asymmetric signing in production with persistent keys and stable `kid` rotation.
- Persist OAuth clients, codes, refresh tokens, revocation state, and replay protection in Mongo.
- Enforce issuer, audience/resource, token purpose, subject, tenant, and scope validation in middleware.
- Harden DCR: redirect URI validation, public vs confidential client separation, hashed client secrets, auditable registration.
- Add logout, logout-all, revoke, refresh rotation, session inventory, and replay detection.
- Fail closed for insecure production config: signing keys, webhook secrets, trusted origins, base URL, and insecure defaults.
- Add SSRF-safe outbound metadata fetching and strict callback/redirect validation.

### Primary File Targets
- `backend/internal/config/config.go`
- `backend/cmd/server/main.go`
- `backend/internal/auth/jwt.go`
- `backend/internal/http/handlers_auth.go`
- `backend/internal/http/handlers_oauth.go`
- `backend/internal/http/middleware.go`
- `backend/internal/http/oauth_state.go`
- `backend/internal/models/models.go`
- `backend/internal/store/store.go`
- `backend/internal/store/mongo.go`
- `backend/internal/store/memory.go`
- `webapp/app/api/session/login/route.ts`
- `webapp/app/api/session/signup/route.ts`
- `webapp/app/api/session/logout/route.ts`
- `webapp/lib/auth-session.ts`
- `webapp/proxy.ts`

### Completion Criteria
- All auth-bearing state is restart-safe and stored durably.
- Browser sessions and MCP OAuth tokens have different purposes and enforcement paths.
- Resource-bound tokens cannot be replayed across unrelated routes or tenants.
- Production boot fails on insecure auth configuration.

## Milestone 2: OAuth Providers And Client Install Interoperability

### Outcome
- Users can sign in with Google and GitHub safely, and MCP clients can complete standards-aligned OAuth install flows.

### Implementation Scope
- Add Google OIDC login with state, nonce, provider claim validation, verified email policy, identity linking, and conflict handling.
- Add GitHub OAuth login with state, verified email fetch, provider subject linking, and collision handling.
- Add explicit account linking and unlinking rules that preserve tenant ownership and do not auto-merge by email.
- Return proper OAuth authorize redirects instead of JSON-only redirect hints.
- Finish public-client-safe DCR and resource-bound authorization behavior for editor and desktop clients.
- Align protected-resource metadata, authorization server metadata, and actual resource URIs used by hub flows.
- Add callback UX and account-link failure UX in the frontend.

### Primary File Targets
- `backend/internal/http/router.go`
- `backend/internal/http/handlers_auth.go`
- `backend/internal/http/handlers_oauth.go`
- `backend/internal/models/models.go`
- `backend/internal/store/store.go`
- `backend/internal/store/mongo.go`
- `webapp/app/login/page.tsx`
- `webapp/app/api/`
- `webapp/lib/api.ts`
- `webapp/lib/auth-session.ts`

### Completion Criteria
- Google and GitHub login work end to end with the same durable session model as credentials login.
- Public MCP clients can discover metadata, register, authorize, exchange tokens, and reauthorize safely.
- Provider identity linking is explicit, auditable, and tenant-safe.

## Milestone 3: Paid Install And Payment Completion

### Outcome
- A buyer can hit a paid install, complete a real checkout/settlement path, receive entitlement, and finish installation without a dead end.

### Implementation Scope
- Introduce a truthful payment capability matrix with runtime-backed statuses: `implemented`, `partial`, `disabled`, `production-ready`.
- Remove incomplete payment methods from production surfaces until their full adapter path exists.
- Complete paid install orchestration: x402 intent creation, provider settlement, entitlement grant, automatic install retry, idempotency, and audit correlation.
- Preserve original x402 requirement/challenge data instead of overwriting it.
- Enforce Stripe webhook signature secrets outside dev mode.
- Harden Stripe Onramp wallet crediting and Stripe Connect readiness/payout gating.
- Make ledger posting, wallet updates, settlement, and payout accounting internally consistent.
- Add buyer `allowedMethods`, merchant per-server `paymentMethods`, and x402 intent history/retry UI.

### Primary File Targets
- `backend/internal/http/handlers_install.go`
- `backend/internal/http/handlers_x402.go`
- `backend/internal/http/x402_service.go`
- `backend/internal/http/payments_helpers.go`
- `backend/internal/http/handlers_payments.go`
- `backend/internal/http/handlers_topups.go`
- `backend/internal/http/handlers_payouts.go`
- `backend/internal/http/accounting_helpers.go`
- `backend/internal/http/stripe_onramp_service.go`
- `backend/internal/http/stripe_connect_service.go`
- `backend/internal/models/models.go`
- `backend/internal/store/memory.go`
- `backend/internal/store/mongo.go`
- `webapp/app/install/[serverId]/page.tsx`
- `webapp/app/buyer/billing/page.tsx`
- `webapp/app/merchant/revenue/page.tsx`
- `webapp/lib/api-client.ts`

### Completion Criteria
- Every enabled payment method has a complete create, verify, settle, and webhook story.
- Paid install completes end to end with no manual backend intervention.
- Webhook paths fail closed in production.
- Buyer and merchant UI never expose unsupported payment flows.

## Milestone 4: Buyer Readiness

### Outcome
- Discovery, install, hub, billing, and recovery flows are coherent, trustworthy, and complete for buyers.

### Implementation Scope
- Replace generic marketplace trust labels with backend-derived signals: hosting mode, compatibility, pricing model, scope preview, verification state, and installability.
- Make the install wizard backend-driven and stateful across refresh/retry.
- Normalize auth, scope, payment, and install recovery responses into a single buyer recovery model.
- Drive client-specific install options from a validated compatibility matrix instead of hardcoded optimism.
- Add buyer hub and local-agent pages with connection health, route projection, scope state, token health, and recovery actions.
- Rework billing to clearly separate wallet balance, top-ups, x402 usage, install-time charges, invoices, and supported payment methods.
- Add post-install verification before marking install complete.

### Primary File Targets
- `webapp/app/marketplace/page.tsx`
- `webapp/app/marketplace/[serverId]/page.tsx`
- `webapp/app/install/[serverId]/page.tsx`
- `webapp/app/api/install/readiness/route.ts`
- `webapp/app/buyer/dashboard/page.tsx`
- `webapp/app/buyer/connections/page.tsx`
- `webapp/app/buyer/billing/page.tsx`
- `webapp/lib/api-client.ts`
- `backend/internal/http/handlers_marketplace.go`
- `backend/internal/http/handlers_install.go`
- `backend/internal/http/handlers_buyer.go`
- `backend/internal/http/handlers_admin.go`
- `backend/internal/models/models.go`

### Completion Criteria
- Buyers can understand what they are buying, how it runs, and what is required before pressing install.
- Install success is reflected in hub, connections, and billing consistently.
- Recovery from auth, scope, or payment failure is explicit and actionable.

## Milestone 5: Merchant Readiness

### Outcome
- Merchants can create/import, validate, deploy, price, publish, observe, and monetize servers without lifecycle loopholes or opaque failures.

### Implementation Scope
- Enforce draft-first creation and remove publish-at-create behavior.
- Add a real Docker import validation pipeline: image reference parsing, tag requirements, registry access validation, manifest/runtime checks, and MCP contract checks.
- Expand deploy state model and make deploy tasks durable, retry-aware, and visible.
- Separate n8n workflow activation from marketplace publishability.
- Add publish gate evaluation with machine-readable blockers.
- Make merchant observability reflect real deploy/auth/payment/install issues.
- Complete merchant payment configuration, pricing readiness, caps, and provider-aware validation.
- Add payout readiness states and clear seller-facing blockers for KYC, holds, tax forms, and admin blocks.

### Primary File Targets
- `backend/internal/http/handlers_merchant.go`
- `backend/internal/http/deploy_worker.go`
- `backend/internal/http/n8n_service.go`
- `backend/internal/http/handlers_business.go`
- `backend/internal/http/handlers_marketplace.go`
- `backend/internal/http/handlers_payments.go`
- `backend/internal/http/handlers_payouts.go`
- `backend/internal/models/models.go`
- `backend/internal/store/store.go`
- `backend/internal/store/mongo.go`
- `webapp/app/merchant/onboarding/page.tsx`
- `webapp/app/merchant/servers/new/import-docker/page.tsx`
- `webapp/app/merchant/servers/[serverId]/deployments/page.tsx`
- `webapp/app/merchant/servers/[serverId]/pricing/page.tsx`
- `webapp/app/merchant/servers/[serverId]/builder/page.tsx`
- `webapp/app/merchant/servers/[serverId]/observability/page.tsx`
- `webapp/app/merchant/revenue/page.tsx`
- `webapp/lib/api-client.ts`

### Completion Criteria
- No server can publish without valid deploy, pricing, payment, and readiness state.
- Merchant surfaces expose real state and no simulated scans, fake deploys, or dead actions.
- Payout readiness and merchant monetization health are explicit.

## Milestone 6: Trust, Operations, And Release Hardening

### Outcome
- Operators can trust what the platform says, diagnose failures quickly, and certify releases with production-like evidence.

### Implementation Scope
- Replace binary verification with evidence-backed verification states for identity, deploy, auth, payments, and operational health.
- Introduce a structured security event taxonomy and enrich audit logs with correlation ID, actor, IP, route, before/after summary, and reason codes.
- Replace static observability with real deploy/auth/payment telemetry.
- Expand readiness beyond storage to include auth metadata, JWKS, webhook secrets, provider configuration, deploy worker health, queue backlog, and publishability.
- Add pagination and scalable query patterns for admin, audit, ledger, payout, and security endpoints.
- Refactor worker coordination for multi-replica safety and add graceful shutdown.
- Add release monitoring tied to runtime diagnostics.

### Primary File Targets
- `backend/internal/http/middleware.go`
- `backend/internal/http/router.go`
- `backend/internal/http/handlers_health.go`
- `backend/internal/http/handlers_admin.go`
- `backend/internal/http/handlers_merchant.go`
- `backend/internal/http/handlers_payments.go`
- `backend/internal/http/handlers_x402.go`
- `backend/internal/http/handlers_payouts.go`
- `backend/internal/http/deploy_worker.go`
- `backend/internal/models/models.go`
- `backend/internal/store/store.go`
- `backend/internal/store/mongo.go`
- `webapp/app/admin/security/page.tsx`
- `webapp/app/admin/audit-logs/page.tsx`
- `webapp/app/admin/payments/page.tsx`
- `webapp/app/admin/client-compatibility/page.tsx`
- `webapp/app/merchant/servers/[serverId]/observability/page.tsx`

### Completion Criteria
- Buyer trust badges and admin diagnostics are backed by stored evidence and telemetry.
- Multi-instance processing is safe for deploy tasks and external event handling.
- Admin and merchant operations surfaces are usable at scale and reflect runtime truth.

## Release Verification Plan

### Backend Gates
- `go test ./...`
- `go vet ./...`
- `go build ./...`
- Startup validation with production-like env config

### Persistence And Runtime Gates
- Mongo-only validation for auth, OAuth, payments, deploy state, audit logs, and security events
- Restart survivability verification
- Duplicate webhook and duplicate settlement replay safety verification

### Frontend Gates
- `npm ci`
- `npm run lint`
- `npm run typecheck`
- `npm run build`
- Clean same-origin session flow verification

### Browser E2E Gates
- Signup, login, logout, MFA continuation
- Google login and GitHub login
- Free install and paid install
- Wallet top-up and settlement retry
- Merchant create/import, deploy, publish, payment config, payout readiness
- Admin diagnostics, payments oversight, security and audit visibility

### Provider Contract Gates
- Stripe Onramp contract tests
- Stripe Connect contract tests
- x402 facilitator verification tests
- n8n deployment contract tests
- OAuth discovery and token endpoint conformance checks

### Compose And Container Gates
- `infra/docker-compose.yml` stack build and boot
- Mongo, backend, webapp, and n8n wiring verification
- Data persistence across restart

### Security Gates
- RBAC and tenant isolation verification
- Secure cookies and no browser-managed bearer tokens
- Webhook secret enforcement
- SSRF-safe readiness metadata fetch
- Dependency and vulnerability scanning

## Sequencing Constraints

- Milestone 1 must finish before Milestone 2 or 3 can be completed safely.
- Milestone 2 and Milestone 3 can overlap only after session and token boundaries are stable.
- Milestone 4 depends on Milestone 2 and Milestone 3 because install recovery and client support rely on both auth and payments.
- Milestone 5 depends on Milestone 3 for truthful monetization and on Milestone 6 for trustworthy operational visibility.
- Release verification runs continuously, but final release gating happens only after all milestone completion criteria are met.

## Definition Of Done

- Auth is production-safe, durable, and resource-bound where required.
- Google and GitHub login are implemented end to end with safe account-linking behavior.
- MCP metadata, hub behavior, install flows, and client compatibility claims are aligned and tested.
- Paid install completes successfully through real provider-backed or wallet-backed settlement flows.
- Buyer, merchant, and admin surfaces expose only truthful, supported actions.
- Trust signals, audit logs, readiness, and telemetry are evidence-backed and operator-usable.
- Backend, frontend, provider, browser, and container verification all pass on the release candidate.
