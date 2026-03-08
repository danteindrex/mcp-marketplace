# Payments Missing Features

This document captures missing payment-system features discovered in the backend/frontend audit.

## P0 (Blockers)

- [ ] Add web checkout flow for paid installs before entitlement check.
  - Problem: install flow currently fails with `entitlement required` for paid servers.
  - Required:
    - Add frontend path to create and settle x402 intents before/within install flow.
    - Handle `402 payment required` and display actionable payment UI.
    - Retry install automatically after successful settlement.
  - Backend endpoints already available:
    - `POST /v1/billing/x402/intents`
    - `POST /v1/billing/x402/intents/{id}/settle`

- [ ] Implement real provider adapters for all advertised payment methods.
  - Problem: `coinbase_commerce` and `stripe` are advertised but not fully implemented as concrete settlement adapters.
  - Required:
    - Add provider-specific session/invoice creation endpoints.
    - Add provider-specific webhook routes and handlers.
    - Map provider settlement payloads to verified x402 settlement records.
    - Prevent methods from being marked as usable when adapter is incomplete.

- [ ] Enforce Stripe webhook signature requirement in production.
  - Problem: onramp webhook verification is bypassed when webhook secret is unset.
  - Required:
    - Reject onramp/connect webhook processing when signature secret is missing in non-dev mode.
    - Add startup/runtime warning if Stripe secret is set but webhook secret is not.
    - Add tests for "secret missing" behavior.

## P1 (High Priority)

- [ ] Add frontend controls to actually select/deselect buyer `allowedMethods`.
  - Problem: buyer page sends `allowedMethods` but does not provide a real selector UX.
  - Required:
    - Method toggles for each supported method.
    - Validation: at least one allowed method remains.
    - Save/restore state from payment controls API.

- [ ] Add frontend controls to edit merchant server `paymentMethods`.
  - Problem: merchant config round-trips `paymentMethods` but UI has no method editor.
  - Required:
    - Method multi-select for per-server method allowlist.
    - Validation against platform-supported method catalog.
    - Clear error handling when server rejects unsupported method.

- [ ] Wire x402 intent lifecycle into web app state/history.
  - Problem: web app has no UI for intent creation/settlement history despite backend support.
  - Required:
    - Intent list and status view.
    - Retry settlement action for pending intents.
    - Tie intent IDs/payment identifiers to install and billing UX.

## P2 (Important, Non-Blocking)

- [ ] Align payment method naming and UX labels.
  - Problem: `stripe_onramp`, `stripe`, `wallet_balance`, and `x402_wallet` can appear inconsistently across controls and records.
  - Required:
    - Define canonical enums and display labels across backend/frontend.
    - Add migration/normalization for existing records where needed.

- [ ] Add operational readiness checks for payment providers.
  - Required:
    - Health endpoint/extensions for payment integrations (x402 facilitator, Stripe Onramp, Stripe Connect, Coinbase adapter when added).
    - Admin surface to show strict configured/not-configured readiness.

- [ ] Add stronger payment-flow observability.
  - Required:
    - Structured audit events for intent create/settle failures.
    - Metrics for settlement failure rate by method/provider.
    - Correlation IDs linking install attempts, intents, and payouts.

## P3 (Quality)

- [ ] Add end-to-end tests that cover frontend-triggered payment flows.
  - Required:
    - Paid install flow (no entitlement -> payment -> entitlement -> install success).
    - Buyer top-up and webhook crediting.
    - Merchant payout profile update + Stripe onboarding + payout run.

- [ ] Add explicit docs for method support matrix.
  - Required:
    - Table mapping each method to: create flow, verify/settle flow, webhook, frontend UX, production readiness.
    - Mark methods as `implemented`, `partial`, or `placeholder`.
