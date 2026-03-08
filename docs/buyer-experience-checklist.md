# Buyer Experience Gap Checklist

Use this list to drive the missing work called out in `requirements.md` and the architecture blueprint. Track each item until we can prove the buyer flow is production-ready.

## 1. One-click installs (CIMD + DCR)
- [x] Wire the frontend install wizard to the real Go `/install` endpoints instead of mock data.
- [x] Implement CIMD fetches that retrieve Authorization Server metadata + JWKS per server.
- [x] Emit proper `WWW-Authenticate` headers (scopes + metadata URLs) from the backend and display them in the wizard.
- [ ] Add integration tests proving Claude Desktop / VS Code can register clients via Dynamic Client Registration.
- [ ] Capture and store issued client metadata so buyers can re-run installs without re-registration.

## 2. Interactive connector previews
- [ ] Stand up Browser MCP (or equivalent) in CI and document how to authenticate it for tests.
- [ ] Build preview components that run inside Claude Desktop, VS Code, Cursor, and Claude Code—each must fetch data live via their MCP APIs.
- [ ] Record fallback behavior for buyers without the extension (static screenshots + instructions).
- [ ] Add smoke tests that open each client, call the preview tools, and verify responses.

## 3. Discovery + ratings data
- [ ] Replace `lib/mock-data.ts` feeds with data pulled from AgentHotspot, Smithery, and our internal moderation queue.
- [ ] Implement the verified 🎖 badge workflow (submission → review → badge assignment) and expose it across listing cards.
- [ ] Surface LOCAL 🏠 vs CLOUD ☁ scopes based on real hosting metadata (remote URI provisioning).
- [ ] Ship rating aggregation (count, average, last updated) and guard against review spam (rate limits + moderation UI).
- [ ] Document how discovery filters map to backend queries so API parity is maintained.

## 4. Install wizard gating + local bridge
- [ ] Make each wizard step (client → auth → scopes → connect → verify) depend on actual backend acknowledgements, not local state.
  - [x] Scopes step now calls `/v1/marketplace/servers/{slug}/scope-check` to confirm entitlements and payments before proceeding.
- [x] Gate the Auth step on successful CIMD/OAuth/JWKS verification with a documented manual override for manual checks.
- [ ] Integrate the Go backend’s scope evaluator so buyers see `insufficient_scope` feedback inline.
- [ ] Implement the `mcp-marketplace://install?...` deep link generator and ensure Codex, Claude Desktop, Cursor, and VS Code open the local bridge on all OS targets.
- [ ] Add regression tests that simulate the bridge flow (CLI auto-accept) and confirm the buyer lands in `/buyer/connections` with a live entitlement.
- [ ] Log every step to the audit service so support can trace failed installs.

## 5. Validation + documentation
- [ ] For every section above, update the testing grid with named suites (Go integration, Python orchestration sims, frontend e2e, MCP client smoke tests).
- [ ] Create buyer-facing docs that explain how one-click install works, including troubleshooting for CIMD, OAuth, and local bridge issues.
- [ ] Ensure `requirements.md` traceability links reference the commits/issues that close each checklist item.

> Status: draft checklist added {{DATE}}. Update this file as tasks complete.
