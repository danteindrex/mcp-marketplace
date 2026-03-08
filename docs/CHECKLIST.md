# Delivery Checklist

## Completed
- [x] Clone provided frontend into monorepo (`webapp/`).
- [x] Build Go backend with buyer/seller/admin role model.
- [x] Implement JWT auth and RBAC middleware.
- [x] Implement marketplace public endpoints.
- [x] Implement buyer endpoints (connections, entitlements, personal hub, local agents).
- [x] Implement merchant endpoints (servers CRUD subset, auth/pricing/observability).
- [x] Implement admin endpoints (tenants, security events, audit logs, entitlement grants).
- [x] Implement x402 v2 intent endpoints (create/list/settle).
- [x] Implement endpoint integration tests.
- [x] Run backend tests (`go test ./...`).
- [x] Run backend build (`go build ./...`).
- [x] Add backend Dockerfile.
- [x] Add webapp Dockerfile.
- [x] Add docker-compose for local full stack.
- [x] Add CI automation workflow.
- [x] Add scheduled security scanning workflow.
- [x] Download free MCP server repository for integration fixtures.
- [x] Require MongoDB connection at startup (no in-memory fallback in normal mode).
- [x] Add n8n service to docker-compose for self-hosted agent builder.
- [x] Add frontend "Agent Builder" entry that opens n8n in a new tab (no iframe).
- [x] Keep deployed servers as marketplace draft until explicit publish action.
- [x] Gate publish on deployed status and non-zero pricing.
- [x] Add persistent deploy queue/outbox model and store APIs.
- [x] Add async deploy worker with retries/backoff and failure handling.
- [x] Expose deploy queue status/errors in merchant deployment APIs.
- [x] Show queue status in merchant deployments UI with auto-refresh while queued.

## Environment Notes
- [x] Docker engine is not available in this Codex runtime, so compose services cannot be started from here.

## Deferred Backlog (Not Part Of This Completed Checklist)
- Add Redis for distributed caching/rate limits/queue coordination.
- Replace static login with full OAuth 2.1 + PKCE + DCR/CIMD implementation.
- Continue wiring remaining webapp views from mock data to backend API calls.
