# API Surface (v1)

Base URL: `/v1`

## Public
- `GET /health`
- `POST /auth/login`
- `GET /v1/marketplace/servers`
- `GET /v1/marketplace/servers/{slug}`

## Authenticated buyer
- `GET /v1/me`
- `GET /v1/buyer/connections`
- `POST /v1/buyer/connections`
- `GET /v1/buyer/entitlements`
- `GET /v1/buyer/hub`
- `GET /v1/buyer/local-agents`
- `POST /v1/buyer/local-agents`
- `GET /v1/billing/x402/intents`
- `POST /v1/billing/x402/intents`
- `POST /v1/billing/x402/intents/{id}/settle`

## Merchant/Admin
- `GET /v1/merchant/servers`
- `POST /v1/merchant/servers`
- `PUT /v1/merchant/servers/{id}`
- `GET /v1/merchant/servers/{id}/observability`
- `GET /v1/merchant/servers/{id}/auth`
- `GET /v1/merchant/servers/{id}/pricing`

## Admin only
- `GET /v1/admin/tenants`
- `GET /v1/admin/security-events`
- `GET /v1/admin/audit-logs`
- `POST /v1/admin/entitlements`

## Test users
- `admin@platform.local`
- `merchant@dataflow.local`
- `buyer@acme.local`