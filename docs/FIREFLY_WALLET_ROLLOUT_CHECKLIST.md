# FireFly Wallet Rollout Checklist

## Pre-Deployment

- [ ] Provision a FireFly Signer instance reachable from the backend.
- [ ] Mount a persistent encrypted directory for `FIREFLY_KEYSTORE_DIR`.
- [ ] Inject `FIREFLY_SIGNER_URL`, `FIREFLY_KEYSTORE_DIR`, and `FIREFLY_KEYSTORE_PASSPHRASE`.
- [ ] Inject `FIREFLY_SIGNER_AUTH_TOKEN` if the signer endpoint is protected.
- [ ] Confirm x402 facilitator settings remain valid.

## Admin Setup

- [ ] Open `/admin/integrations`.
- [ ] Enable `Keep FireFly available`.
- [ ] Select `firefly` as the managed wallet provider.
- [ ] Keep `Managed wallet auto-pay` enabled for the default marketplace path.
- [ ] Decide whether `Legacy payment mode` stays enabled during rollout.
- [ ] Decide whether `Expose external wallet mode` should remain hidden.
- [ ] Save and verify the wallet section reports the expected active provider.

## Validation

- [ ] Buyer billing shows the expected backend and managed wallet address.
- [ ] Merchant revenue shows the expected receive-wallet backend.
- [ ] A paid install can create and settle an x402 intent.
- [ ] An MCP hub paid tool call auto-pays successfully.
- [ ] Stored x402 intents include wallet provider and wallet address attribution.

## Migration

- [ ] Decide whether existing managed-wallet records should stay on CDP until first reprovision or be reprovisioned immediately.
- [ ] Review seller server payment addresses before switching revenue intake for live listings.
- [ ] Communicate any wallet-address change risk to sellers before changing the default backend in production.

## Rollback

- [ ] Re-open `/admin/integrations`.
- [ ] Switch `wallet.provider` back to `cdp`.
- [ ] Leave `cdpEnabled=true`.
- [ ] Save and verify buyer and merchant surfaces return to the CDP backend.
