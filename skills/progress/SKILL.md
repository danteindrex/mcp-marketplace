---
name: progress
description: Use when continuing work on the local MCP Marketplace repo and you need a compact handoff of what has already been completed, what is currently broken, and what still needs live verification.
---

# Progress

Use this skill when resuming work on `C:\Users\user\Documents\mcp-marketplace`.

## Completed

- Backend dependency metadata was fixed and Docker backend builds succeed.
- Host/backend port conflict was moved off `8080`; backend now runs on host `2024`.
- Integration settings are persisted through the admin UI and mirrored into [`/.env`](C:/Users/user/Documents/mcp-marketplace/.env).
- Configured integrations already saved in UI:
  - Google OAuth
  - GitHub OAuth
  - Stripe keys and webhook secrets
  - x402 testnet facilitator
  - n8n base URL and API key
- Super admin seed credentials:
  - `admin@platform.local`
  - `ChangeMe123!dev`
- Mock/demo identity fallbacks like `user@example.com` and `${role}@local` were removed from the main shell/sidebar.
- Logout behavior was fixed to clear session state and redirect reliably.
- Marketplace/home auth CTAs were improved so logged-in users do not see stale login buttons after auth resolves.
- Merchant import/pricing flow was made real:
  - Docker image import no longer relies on fake scan output
  - pricing page supports real price entry and publish flow
- Real container deployment path was added:
  - `local-docker` deploy target
  - server model stores `containerPort` and runtime container id
  - backend deploy worker now runs the submitted image through Docker and persists the runtime URL
- Real deployment proof already captured:
  - a public Docker image was deployed
  - the assigned runtime URL returned `HTTP 200`
- Agent Builder backend persistence now exists:
  - `GET/PUT /v1/merchant/servers/{id}/builder`
  - builder config is stored on the server record
- Missing merchant Builder UI page was added at:
  - [`webapp/app/merchant/servers/[serverId]/builder/page.tsx`](C:/Users/user/Documents/mcp-marketplace/webapp/app/merchant/servers/[serverId]/builder/page.tsx)

## Confirmed External Constraints

- Stripe Connect code path is structurally correct, but the current Stripe account is not enrolled as a Connect platform.
- Stripe Onramp endpoint is also blocked by Stripe account capability/access on this account.
- These provider-side blockers cannot be forced to return `200` by app code alone.

## Current Breakers

- If `http://localhost:3000` fails with empty response while backend health is fine, check webapp port wiring first.
- Current known cause:
  - webapp inherited `PORT=2024`
  - Next started inside the container on `2024`
  - compose still exposed `3000:3000`
  - result: `localhost:3000` returned empty response

## What Still Needs Live Verification

- Webapp reachable again on `http://localhost:3000`
- Merchant Builder page loads and saves round-trip
- Seller flow on a fresh server:
  - create/import
  - builder save
  - deploy
  - pricing save
  - publish
- Buyer flow on the published server:
  - marketplace listing
  - detail page
  - install
  - connection appears in buyer connections
- OAuth browser completion:
  - Google login full callback completion
  - GitHub login full callback completion
- Stripe UI behavior after provider blockers:
  - verify app surfaces real provider capability errors instead of generic failures

## Fast Resume Checks

1. Backend health:
   - `curl.exe -i http://localhost:2024/health`
2. Webapp root:
   - `curl.exe -i http://localhost:3000`
3. Merchant Builder route exists in build output:
   - `/merchant/servers/[serverId]/builder`
4. If browser tabs are stale after container restart, reopen fresh tabs instead of trusting `chrome-error://chromewebdata/`.
