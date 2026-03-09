# MCP Marketplace Monorepo

## Repository layout
- `backend/`: Go API service (auth, marketplace, buyer, merchant, admin, OAuth, x402).
- `webapp/`: Next.js frontend wired to backend APIs.
- `infra/docker-compose.yml`: local backend + webapp + mongo + n8n composition.
- `docs/`: API and implementation notes.
- `mcp-servers/modelcontextprotocol-servers/`: external MCP server reference snapshot.

## Prerequisites
- Go `1.26+`
- Node `22+`
- MongoDB `7+` (or Docker)

## Local development

### Backend
```bash
cd backend
# ensure MongoDB is reachable (default: mongodb://localhost:27017)
export MONGO_URI="mongodb://localhost:27017"
export MONGO_DB_NAME="mcp_marketplace"
export MONGO_REQUIRED="true"
export N8N_BASE_URL="http://localhost:5678"
export N8N_API_KEY="<n8n-api-key>" # optional but recommended for auto deploy
export N8N_TIMEOUT_SECONDS="12"
export SUPER_ADMIN_EMAIL="admin@platform.local"
export SUPER_ADMIN_PASSWORD="<strong-password>"
export JWT_SECRET="<strong-jwt-secret>"
export MCP_SDK_ENABLED="true" # optional: enable MCP SDK-backed initialize/tools/list handling
# in-memory fallback is disabled by design

# x402 + payment rails
export X402_MODE="facilitator" # or "mock" for local testing
export X402_FACILITATOR_URL="https://<your-facilitator>"
export X402_FACILITATOR_API_KEY="<facilitator-key>"
export SUPPORTED_PAYMENT_METHODS="x402_wallet,wallet_balance,stripe"

# Stripe Onramp (card/ACH -> USDC prepaid wallet)
export STRIPE_SECRET_KEY="sk_live_..."
export STRIPE_WEBHOOK_SECRET="whsec_..."
export STRIPE_ONRAMP_MIN_USD="10"
export STRIPE_ONRAMP_DEFAULT_USD="50"
export STRIPE_ONRAMP_RETURN_URL="http://localhost:3000/buyer/billing"
export STRIPE_ONRAMP_REFRESH_URL="http://localhost:3000/buyer/billing"

# seed super admin + start server (PowerShell)
./scripts/start-with-seed.ps1

# seed super admin + start server (bash)
./scripts/start-with-seed.sh

# one-time Windows protocol setup for one-click local installs
./scripts/install-local-bridge.ps1

go test ./...
go run ./cmd/server
```

### Webapp
```bash
cd webapp
npm ci
export NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY="pk_live_..."
export NEXT_PUBLIC_N8N_URL="http://localhost:5678"
npm run lint
npm run typecheck
npm run dev
```

## Session model
- Web login/signup goes through Next API routes:
  - `POST /api/session/login`
  - `POST /api/session/signup`
  - `POST /api/session/logout`
- Access token is stored in an `HttpOnly` cookie (`mcp_access_token`).
- Browser code no longer persists access tokens in `localStorage`.

## CI
- `.github/workflows/ci.yml`
  - backend: `go test`, `go vet`, `go build`
  - webapp: `npm ci`, `npm run lint`, `npm run typecheck`, `npm run build`
- `.github/workflows/security.yml`
  - weekly `govulncheck`
