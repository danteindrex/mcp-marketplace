# FireFly Managed Wallet Architecture

## Goal

Add a vendor-independent managed-wallet backend for x402 payments without removing the existing wallet or payment implementations.

## Architecture

- The marketplace application remains the wallet product layer.
- Hyperledger FireFly Signer is the signing engine for the FireFly backend.
- FireFly-compatible Keystore V3 files remain on disk in the configured keystore directory.
- The marketplace database stores ownership, policy, provider selection, attribution, and spend controls.
- Existing CDP and wallet-balance paths stay intact and remain toggleable.

## Backend Selection

Admin integrations now control:

- `wallet.provider`: preferred managed-wallet backend
- `wallet.managedAutoPayEnabled`: whether the marketplace hub can auto-sign x402 payments
- `wallet.legacyPaymentModeEnabled`: whether legacy payment methods such as `wallet_balance` remain selectable
- `wallet.externalWalletsEnabled`: whether advanced external-wallet UI can be exposed
- `wallet.cdpEnabled`: whether the current CDP backend remains selectable
- `wallet.fireflyEnabled`: whether FireFly remains selectable

If the selected provider is disabled, the backend falls back to another enabled provider instead of breaking existing flows.

## FireFly Storage Model

- `FIREFLY_KEYSTORE_DIR` points to the signer-visible keystore directory.
- New FireFly wallets are created as Keystore V3 files.
- The app normalizes files to `<address>.key.json` plus `<address>.password` so FireFly fswallet-style directory scanning can consume them.
- Secrets such as `FIREFLY_KEYSTORE_PASSPHRASE` and `FIREFLY_SIGNER_AUTH_TOKEN` should come from secret management, not source control.

## Payment Execution

- Buyer managed wallets remain one wallet per marketplace user.
- Seller managed wallets remain one default receive wallet per seller tenant.
- Install-time x402 flow still creates intents and can auto-settle with `x402_wallet` when auto-pay is enabled.
- MCP hub per-call flow still auto-pays through the marketplace hub.
- `wallet_balance` still works when legacy mode stays enabled.

## Attribution

Each x402 intent now preserves:

- `tenantId`
- `userId`
- `serverId`
- `toolName`
- `paymentIdentifier`
- `walletId`
- `walletAddress`
- `walletProvider`

This keeps autonomous account-bound payments attributable to the authenticated marketplace user and the managed wallet backend used.

## Security Assumptions

- FireFly signer and the app share access to the keystore directory.
- The keystore directory must be mounted on encrypted storage with OS-level access restrictions.
- The passphrase should be injected from a secret store and rotated operationally.
- `ALLOW_INSECURE_DEFAULTS=true` only exists for local and test fallback behavior.
- Facilitator verification still governs payment settlement truth in production.

## Deployment Notes

- Set the FireFly env vars in the platform environment or save them through `/admin/integrations`.
- Mount `FIREFLY_KEYSTORE_DIR` into both the backend and the FireFly Signer runtime if you run them as separate services.
- Keep the admin-saved config as the primary source; env remains fallback only.
