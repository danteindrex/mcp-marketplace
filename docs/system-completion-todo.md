# System Completion TODO

Track platform setup and end-to-end verification here. This file assumes credentials are entered manually by the account owner in the Admin Integrations UI.

## 1. Credentials To Collect

- [ ] Google OAuth client ID
- [ ] Google OAuth client secret
- [ ] Google OAuth redirect base
- [ ] GitHub OAuth client ID
- [ ] GitHub OAuth client secret
- [ ] GitHub OAuth redirect base
- [ ] Stripe publishable key
- [ ] Stripe secret key
- [ ] Stripe webhook secret
- [ ] Stripe Connect webhook secret
- [ ] Stripe onramp return URL
- [ ] Stripe onramp refresh URL
- [ ] Stripe Connect return URL
- [ ] Stripe Connect refresh URL
- [ ] x402 mode selected
- [ ] x402 facilitator URL
- [ ] x402 facilitator API key
- [ ] n8n base URL
- [ ] n8n API key

## 2. Where To Get Them

- [ ] Stripe
  - API keys come from the Stripe Developers Dashboard API keys page.
  - Webhook secrets come from Stripe webhook endpoint configuration.
  - Source: https://docs.stripe.com/keys
- [ ] Google OAuth
  - Create a Web application OAuth client in Google Auth platform / Google Cloud credentials.
  - Source: https://developers.google.com/workspace/guides/create-credentials
- [ ] GitHub OAuth
  - Create a GitHub OAuth App under Settings > Developer settings > OAuth Apps.
  - Source: https://docs.github.com/en/developers/apps/creating-an-oauth-app
- [ ] n8n
  - Create the API key in Settings > n8n API.
  - Source: https://docs.n8n.io/api/authentication/

## 3. Enter In UI

- [ ] Log in as platform admin
- [ ] Open `/admin/integrations`
- [ ] Save Google OAuth settings
- [ ] Save GitHub OAuth settings
- [ ] Save Stripe settings
- [ ] Save x402 settings
- [ ] Save n8n settings
- [ ] Refresh the page and confirm each section shows `Configured`

## 4. Verify Platform Flows

- [ ] Password login succeeds
- [ ] Logout returns to `/login`
- [ ] Home page detects session and shows `Dashboard`
- [ ] Marketplace page detects session and does not show stale `Login`
- [ ] Google OAuth login works
- [ ] GitHub OAuth login works
- [ ] Buyer billing page loads with runtime Stripe publishable key
- [ ] Stripe onramp session can be created
- [ ] Agent Builder opens configured n8n URL
- [ ] Merchant server creation works
- [ ] Merchant deployment queue works with n8n configured
- [ ] MCP marketplace install flow works from UI
- [ ] Admin integrations persist after page refresh

## 5. Known Constraints

- [ ] Secrets must be collected and pasted by the account owner
- [ ] External account login and secret retrieval are not automated here
- [ ] Full payment verification needs real provider accounts and callbacks

