# Webapp Status

## Current state
- Frontend pages are wired to backend API endpoints through `lib/api.ts` and `lib/api-client.ts`.
- Auth/session flow is cookie-session based:
  - Login/signup via `app/api/session/*`.
  - `mcp_access_token` is `HttpOnly`.
- Route protection runs in `proxy.ts`.
- Merchant, buyer, and admin pages are present in the App Router.

## Quality gates
- `npm run lint` is configured and passing (warnings only).
- `npm run typecheck` is passing.
- `next build` compiles successfully in this environment but exits with `spawn EPERM` during final process execution.

## Known gaps
- Some UI code still has lint warnings (unused vars / hook dependency warnings).
- Backend is still in-memory/persisted-file data, not production datastore.
- OAuth/x402 implementations are functional but intentionally lightweight.

