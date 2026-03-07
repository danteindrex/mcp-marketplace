# API Surface (v1)

Base URL: `http://localhost:8080`

## Authentication model
- `POST /auth/signup` and `POST /auth/login` return JWT bearer tokens.
- Protected routes accept either:
  - `Authorization: Bearer <token>`, or
  - `mcp_access_token` cookie.
- OAuth authorization endpoint requires authenticated caller and enforces resource subject binding.

## Public routes
- `GET /health`
- `GET /healthz`
- `POST /auth/signup`
- `POST /auth/login`
- `POST /webhooks/stripe/onramp`
- `POST /webhooks/stripe/connect`
- `GET /.well-known/oauth-protected-resource`
- `GET /.well-known/oauth-authorization-server`
- `POST /oauth/register`
- `POST /oauth/token`

## Authenticated OAuth route
- `GET /oauth/authorize`
- `GET /mcp/hub/{tenantID}/{userID}`
- `POST /mcp/hub/{tenantID}/{userID}`

## Public marketplace (`/v1`)
- `GET /v1/marketplace/servers`
- `GET /v1/marketplace/servers/{slug}`

## Authenticated routes (`/v1`)
- `GET /v1/me`
- `POST /v1/marketplace/servers/{slug}/install`

### Settings
- `GET /v1/settings/profile`
- `PUT /v1/settings/profile`
- `PUT /v1/settings/security/password`
- `GET /v1/settings/preferences`
- `PUT /v1/settings/preferences`
- `GET /v1/settings/notifications`
- `PUT /v1/settings/notifications`

### Buyer
- `GET /v1/buyer/connections`
- `POST /v1/buyer/connections`
- `POST /v1/buyer/connections/{id}/rotate`
- `POST /v1/buyer/connections/{id}/revoke`
- `GET /v1/buyer/entitlements`
- `GET /v1/buyer/billing`
- `GET /v1/buyer/invoices`
- `GET /v1/buyer/hub`
- `GET /v1/buyer/payments/controls`
- `PUT /v1/buyer/payments/controls`
- `GET /v1/buyer/payments/topups`
- `POST /v1/buyer/payments/topups/stripe/session`
- `GET /v1/buyer/local-agents`
- `POST /v1/buyer/local-agents`

### x402 billing
- `GET /v1/billing/x402/intents`
- `POST /v1/billing/x402/intents`
- `POST /v1/billing/x402/intents/{id}/settle`

## Merchant + admin routes
- `GET /v1/merchant/servers`
- `POST /v1/merchant/servers`
- `GET /v1/merchant/servers/{id}`
- `PUT /v1/merchant/servers/{id}`
- `POST /v1/merchant/servers/{id}/deploy`
- `POST /v1/merchant/servers/{id}/publish`
- `GET /v1/merchant/revenue`
- `GET /v1/merchant/servers/{id}/observability`
- `GET /v1/merchant/servers/{id}/auth`
- `GET /v1/merchant/servers/{id}/pricing`
- `GET /v1/merchant/servers/{id}/deployments`
- `GET /v1/merchant/servers/{id}/builder`
- `GET /v1/merchant/payments/overview`
- `GET /v1/merchant/payments/payout-profile`
- `PUT /v1/merchant/payments/payout-profile`
- `POST /v1/merchant/payments/payout-profile/stripe/onboarding-link`
- `POST /v1/merchant/payments/payout-profile/stripe/refresh-kyc`
- `GET /v1/merchant/payments/ledger`
- `GET /v1/merchant/payments/payouts`
- `GET /v1/merchant/servers/{id}/payments/config`
- `PUT /v1/merchant/servers/{id}/payments/config`

### Merchant lifecycle notes
- New servers default to marketplace `draft` and deployment `not_deployed`.
- `POST /deploy` keeps listing in draft and marks deployment as active.
- `POST /publish` requires deployment to be active and `pricingAmount > 0`.
- Public marketplace endpoints only return `published` listings.

## Admin-only routes
- `GET /v1/admin/tenants`
- `GET /v1/admin/security-events`
- `GET /v1/admin/audit-logs`
- `GET /v1/admin/client-compatibility`
- `GET /v1/admin/payments/overview`
- `GET /v1/admin/payments/fee-policies`
- `PUT /v1/admin/payments/fee-policies`
- `GET /v1/admin/payments/ledger`
- `GET /v1/admin/payments/reconciliation`
- `GET /v1/admin/payments/payout-profiles`
- `PUT /v1/admin/payments/payout-profiles/{tenantID}/block`
- `GET /v1/admin/payments/payouts`
- `POST /v1/admin/payments/payouts/run`
- `POST /v1/admin/entitlements`
