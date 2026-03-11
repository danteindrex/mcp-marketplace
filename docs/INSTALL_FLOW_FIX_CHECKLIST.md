# Install Flow Fix Checklist

- [x] Make install auth readiness validate marketplace-owned metadata instead of upstream runtime URLs.
- [x] Add a marketplace `/.well-known/mcp.json` endpoint that links to the existing OAuth metadata and JWKS.
- [x] Expose discovery URLs from the marketplace server detail payload so the frontend can use a stable contract.
- [x] Stop auto-seeding placeholder `https://example.com/...` canonical resource URLs in the merchant create flow.
- [x] Allow draft server creation without a pre-filled canonical runtime URL; deployment can populate the upstream runtime later.
- [x] Add backend regression coverage for marketplace metadata and install detail discovery URLs.
- [x] Add publish-time validation that distinguishes buyer-facing install metadata from upstream runtime targets more explicitly.
- [x] Remove hardcoded frontend backend URLs so Docker Compose env stays the single API-base source.
- [ ] Add frontend automated tests for the install wizard readiness state once a frontend test harness exists.
