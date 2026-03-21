# Unfinished Features Checklist

This checklist is based on the current code state as of March 21, 2026.

It intentionally excludes work that is already materially implemented in code today, including:
- buyer payment method controls
- buyer x402 intent history
- merchant per-server payment method editing
- install wizard wiring to backend APIs
- marketplace-owned MCP/OAuth metadata endpoints
- managed wallet auto-settle paths for `wallet_balance` and `x402_wallet`

The remaining work below is what still appears unfinished, incomplete, or not yet hardened for release.

## Phase 1: Core Release Blockers

### Auth and OAuth hardening
- [ ] Replace the current 8-hour bearer-cookie session model with durable session and refresh-token flows.
- [ ] Add session inventory, revoke, logout-all, and replay detection for browser sessions.
- [ ] Change OAuth authorize to support standards-aligned redirect behavior instead of JSON-only redirect payloads.
- [ ] Hash confidential OAuth client secrets at rest instead of storing raw values.
- [ ] Add stricter DCR validation for redirect URIs, public vs confidential clients, and registration auditability.
- [ ] Persist and reuse client registrations for repeat installs where appropriate.

### Paid install and payment completeness
- [ ] Finish provider-backed payment adapters beyond wallet flows.
- [ ] Implement direct Stripe x402 settlement only when create, verify, settle, and webhook handling are complete.
- [ ] Implement Coinbase Commerce only when create, verify, settle, and webhook handling are complete.
- [ ] Keep unsupported payment methods hidden or disabled everywhere until adapters are complete.
- [ ] Add clear install retry and recovery handling for non-wallet payment methods.
- [ ] Verify all payment method names and labels are normalized across backend, frontend, ledger, and payout records.

### Install interoperability
- [ ] Add end-to-end tests for Claude Desktop, VS Code, Cursor, Codex, and local bridge acceptance flows.
- [ ] Add regression coverage for install wizard `402` recovery and automatic post-payment retry.
- [ ] Verify repeat install behavior when an OAuth client was already registered previously.
- [ ] Surface stronger troubleshooting states for DCR, PKCE, resource mismatch, and insufficient-scope failures.

### Deployment truthfulness
- [ ] Remove or strictly isolate insecure-default deployment shortcuts from release paths.
- [ ] Ensure deployment success always reflects a real provider action, not just optimistic local status mutation.
- [ ] Add restart-safe, multi-instance-safe deploy worker coordination.
- [ ] Add operational visibility for queue backlog, retry counts, failure reasons, and worker health.

## Phase 2: Product Completion

### Buyer experience
- [ ] Add stronger post-install verification before marking install complete.
- [ ] Add a clearer recovery model across auth failures, payment failures, scope failures, and bridge failures.
- [ ] Finish local agent and buyer hub operational status views so connections show real health and actionable remediation.
- [ ] Replace any remaining hardcoded client optimism with a validated compatibility matrix.

### Merchant experience
- [ ] Harden Docker import validation beyond basic form acceptance.
- [ ] Add image/reference validation, registry access checks, manifest/runtime checks, and MCP contract checks.
- [ ] Improve merchant observability so it reports real runtime, auth, deploy, and payment failures instead of sparse aggregates.
- [ ] Add clearer payout readiness blockers for KYC, holds, tax status, and admin blocks.

### Admin and trust surfaces
- [ ] Expand readiness diagnostics for auth, JWKS, Stripe, x402 facilitator, wallet provider, n8n, and deployment workers.
- [ ] Improve audit/security event detail with better reason codes and correlation IDs.
- [ ] Add payment provider health and readiness visibility in admin UI.
- [ ] Make verification and trust badges evidence-backed instead of inferred from static listing fields.

## Phase 3: Production Hardening

### Testing and release gates
- [ ] Get `go test ./...` consistently passing in a clean environment with Mongo-backed verification where needed.
- [ ] Add frontend/browser E2E coverage for install, wallet top-up, payout onboarding, and admin integration flows.
- [ ] Add duplicate webhook, duplicate settlement, and replay-protection tests.
- [ ] Add Compose-level verification for mongo, backend, webapp, and n8n together.

### Persistence and scale
- [ ] Verify restart survivability for sessions, OAuth state, payments, deploy tasks, and hub projections.
- [ ] Add pagination and scalable querying for audit logs, payouts, ledger, and admin listings.
- [ ] Add stronger metrics for payment failure rate, deploy failure rate, and install conversion.
- [ ] Add structured correlation between install attempts, intents, settlements, ledger entries, and payouts.

### Docs and status hygiene
- [ ] Reconcile older docs with the actual code state so checklists stop marking completed work as missing.
- [ ] Create a payment support matrix that explicitly marks each method as implemented, partial, disabled, or placeholder.
- [ ] Document exact release prerequisites for production mode, including JWT keys, webhook secrets, wallet provider credentials, and Mongo requirements.

## Final Phase: ChatGPT App Support

Do this only after the phases above are complete enough that the marketplace can truthfully expose stable public MCP experiences.

### Why this is separate
- [ ] Keep ChatGPT app work out of the critical path until auth, payments, install reliability, and deployment truthfulness are release-ready.
- [ ] Treat ChatGPT app support as an additional distribution surface on top of a stable MCP marketplace, not as a substitute for fixing core flows.

### What official OpenAI docs imply for this repo
- [ ] Support a public HTTPS MCP endpoint suitable for ChatGPT connector creation, not only local/dev endpoints.
- [ ] Add first-class support for ChatGPT Apps UI resources, not just tool-only MCP connectors.
- [ ] Model UI resources around the MCP Apps standard using `_meta.ui.resourceUri`.
- [ ] Support embedded app UI flows that communicate over the standard `ui/*` bridge.
- [ ] Use ChatGPT-specific `window.openai` extensions only when needed, with graceful fallback behavior.
- [ ] Ensure tool metadata is high quality enough for ChatGPT discovery and launcher ranking.
- [ ] Add content security policy handling for any app UI that fetches external resources.

### Repo work needed for ChatGPT app support
- [ ] Decide whether ChatGPT app support is marketplace-wide or per-server capability.
- [ ] Extend server/listing models to mark `supportsChatGPTApp` separately from plain ChatGPT connector support.
- [ ] Add merchant configuration for ChatGPT app metadata:
- [ ] app display name
- [ ] app description
- [ ] logo/icon assets
- [ ] screenshots
- [ ] privacy policy URL
- [ ] company/publisher URL
- [ ] localization fields
- [ ] Add support for MCP UI resource registration and storage for each app-capable server.
- [ ] Add a frontend surface for merchants to build or upload ChatGPT app UI resources.
- [ ] Add testing flows for embedded app UI in ChatGPT, including mobile behavior if custom UI is used.
- [ ] Add app-specific review readiness checks before allowing a merchant to mark a server as ChatGPT-app-capable.
- [ ] Add admin moderation/review support for ChatGPT app metadata and screenshots.

### ChatGPT testing and submission checklist
- [ ] Verify the MCP server is reachable over public HTTPS.
- [ ] Verify connector creation works in ChatGPT developer mode.
- [ ] Verify tool discovery and invocation from ChatGPT.
- [ ] Verify read-only vs write/destructive annotations are accurate.
- [ ] Verify any write actions require the right confirmation behavior.
- [ ] Verify screenshots and descriptions match actual functionality.
- [ ] Verify the app is complete, stable, low-latency, and not a demo-only surface.
- [ ] Verify organization identity and owner-role submission prerequisites.
- [ ] Verify CSP and security/privacy requirements before submission.

## Official OpenAI References

These are the sources used for the ChatGPT app phase:
- Apps SDK overview: https://developers.openai.com/
- MCP Apps in ChatGPT: https://developers.openai.com/apps-sdk/mcp-apps-in-chatgpt
- MCP and Apps SDK concepts: https://developers.openai.com/apps-sdk/concepts/mcp-server
- Connect from ChatGPT: https://developers.openai.com/apps-sdk/deploy/connect-chatgpt
- Submit and maintain your app: https://developers.openai.com/apps-sdk/deploy/submission
- App submission guidelines: https://developers.openai.com/apps-sdk/app-submission-guidelines
