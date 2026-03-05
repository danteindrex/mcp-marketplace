# MCP Marketplace Monorepo

## Structure
- `backend/` Go API backend.
- `webapp/` Next.js frontend (cloned/integrated from provided repo).
- `infra/docker-compose.yml` local orchestration.
- `mcp-servers/modelcontextprotocol-servers` free MCP server references.
- `docs/` implementation artifacts.

## Quick start
### Backend
1. Install Go 1.26+
2. Run:
   - `cd backend`
   - `go test ./...`
   - `go run ./cmd/server`

### Webapp
1. Install Node 22+
2. Run:
   - `cd webapp`
   - `npm install`
   - `npm run build`
   - `npm run dev`

## CI automations
- `.github/workflows/ci.yml` (build + tests)
- `.github/workflows/security.yml` (weekly `govulncheck`)