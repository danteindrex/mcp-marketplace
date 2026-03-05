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

## Environment limitations encountered
- [ ] Docker engine not installed locally, so image build/run not executed in this environment.

## Next production actions
- [ ] Persist data in MongoDB/Redis.
- [ ] Replace static login with full OAuth 2.1 + PKCE + DCR/CIMD implementation.
- [ ] Wire webapp pages from mock data to backend API calls.