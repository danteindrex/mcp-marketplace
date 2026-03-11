# MCP Marketplace Webapp

## Run locally
```bash
npm ci
npm run lint
npm run typecheck
npm run dev
```

Default app URL: `http://localhost:3000`

## Backend integration
- Configure API base in `.env.local`:
  - `NEXT_PUBLIC_API_BASE_URL=<your-backend-base-url>`
  - `NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_test_...` (required for embedded Stripe Onramp widget)
- Frontend calls backend via `lib/api.ts` and `lib/api-client.ts`.

## Session flow
- Login/signup requests hit:
  - `POST /api/session/login`
  - `POST /api/session/signup`
- These routes call backend auth endpoints and set:
  - `mcp_access_token` (`HttpOnly`)
  - `mcp_active_role`
- Logout:
  - `POST /api/session/logout`

## Scripts
- `npm run dev`: start dev server
- `npm run lint`: run ESLint
- `npm run typecheck`: run TypeScript check
- `npm run build`: production build
- `npm run start`: run production server
