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
export X402_MODE="facilitator" # or "disabled"/"test" for local testing
export X402_FACILITATOR_URL="https://<your-facilitator>"
export X402_FACILITATOR_API_KEY="<facilitator-key>"
export SUPPORTED_PAYMENT_METHODS="x402_wallet,wallet_balance,stripe"

# Managed marketplace wallets (provider-managed signing for seamless x402)
export WALLET_PROVIDER="cdp"
export WALLET_MANAGED_AUTOPAY_ENABLED="true"
export WALLET_LEGACY_PAYMENT_MODE_ENABLED="true"
export WALLET_EXTERNAL_WALLETS_ENABLED="false"
export WALLET_CDP_ENABLED="true"
export WALLET_FIREFLY_ENABLED="false"
export CDP_API_KEY_ID="<cdp-api-key-id>"
export CDP_API_KEY_SECRET="<cdp-api-key-secret>"
export CDP_WALLET_SECRET="<cdp-wallet-secret>"
export FIREFLY_SIGNER_URL="http://localhost:8545" # when WALLET_PROVIDER=firefly
export FIREFLY_SIGNER_AUTH_TOKEN=""
export FIREFLY_KEYSTORE_DIR="./data/firefly/keystore"
export FIREFLY_KEYSTORE_PASSPHRASE="<strong-passphrase>"
export WALLET_DEFAULT_NETWORK="base"
export WALLET_DEFAULT_ASSET="USDC"
export WALLET_CUSTODY_MODE="provider_managed"

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

## Wallet docs
- `docs/FIREFLY_MANAGED_WALLETS.md`
- `docs/FIREFLY_WALLET_ROLLOUT_CHECKLIST.md`
- `docs/WALLET_IMPLEMENTATION_CHECKLIST.md`
