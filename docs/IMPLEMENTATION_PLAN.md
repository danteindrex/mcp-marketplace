# Backend Plan (Go-First)

## Roles and tenancy
- Buyer/User: owns personal hub endpoint and allowed endpoints.
- Seller/Tenant: publishes servers, configures pricing/auth/deploy.
- Admin/Owner: global governance, security, entitlements.

## Services
- Go API Gateway + REST backend (implemented in this repo).
- Go auth token minting (JWT in v1 implementation with role + tenant claims).
- Go x402 v2 payment intents (challenge + settlement endpoints).
- Personal MCP Hub projection (single hub route catalog for each user).
- Local bridge registration endpoints for local MCP servers.

## Data approach
- Current implementation: in-memory store with deterministic seed data.
- Production target: MongoDB + Redis + ClickHouse + Kafka/NATS.

## Security controls implemented
- JWT auth middleware.
- Tenant boundary checks.
- Role-based route guards (buyer/merchant/admin).
- Security headers middleware.
- Audit logging on mutating operations.

## Deployment strategy
- Multi-region ready API stateless service.
- Docker image for backend and webapp.
- CI: build + tests for backend and webapp.
- Weekly govulncheck automation.

## Gaps to close for production hardening
- Replace in-memory store with Mongo/Redis adapter.
- Replace dev JWT secret handling with KMS.
- Add token introspection/revocation endpoint.
- Add distributed rate limits and WAF rules.
- Add OpenTelemetry traces and SIEM shipping.