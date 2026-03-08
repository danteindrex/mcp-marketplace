# x402 Payments Skill

Use this skill for paid installs, wallet flows, HTTP 402 behavior, micropayments, and settlement orchestration.

## Why This Skill Exists

Research notes:
- x402 is an HTTP-native payment model that uses `402 Payment Required` as part of the request/response flow.
- A typical x402 flow is: request -> `402` with payment requirements -> signed payment payload -> retried request -> settlement confirmation.
- The protocol is explicitly positioned for APIs, apps, and AI agents using stablecoin-based payments.

## What This Repo Needs

- a complete paid install flow instead of entitlement dead-ends
- x402 intent creation, settlement, retry, and auditability
- alignment between wallet balances, provider top-ups, and merchant revenue events
- clear UX and API contracts for buyers, merchants, and admins

## Decision Rules

- a paid install is not complete until payment leads to entitlement and successful install retry
- `402` responses must be actionable and structured, not generic failures
- do not advertise payment methods that do not have a real create + verify + settle path
- tie settlement records to install attempts and user-visible outcomes
- keep method names canonical across backend models, APIs, and frontend labels

## Implementation Focus

- x402 intent endpoints and lifecycle
- payment-assisted install flow
- wallet top-ups, balance usage, and settlement records
- Stripe onramp and payout system interactions
- merchant pricing configuration and buyer method controls
- payment observability, webhook correctness, and reconciliation

## Done Looks Like

- buyers can complete paid installs in one guided flow
- failed settlement states are recoverable and observable
- payment and install records correlate cleanly
- method availability matches real implementation readiness
