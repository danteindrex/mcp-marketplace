# Managed Wallet Checklist

## Default Managed Wallet Path

- [x] Add first-class managed wallet model for buyers and sellers.
- [x] Auto-provision buyer managed wallets.
- [x] Auto-provision seller managed wallets.
- [x] Default seller server `paymentAddress` from seller managed wallet.
- [x] Remove manual x402 proof entry from the normal install flow.
- [x] Auto-settle `x402_wallet` payments during install.
- [x] Auto-settle `x402_wallet` payments for MCP tool calls through the hub.
- [x] Keep payer identity bound to authenticated marketplace hub user context.
- [x] Surface buyer managed wallet in billing UI.
- [x] Surface seller managed wallet in merchant revenue UI.

## Admin Configuration

- [x] Keep env config as fallback only.
- [x] Add wallet provider settings to admin integrations backend.
- [x] Add wallet provider settings to admin integrations UI.
- [x] Expose runtime wallet metadata for the frontend.

## CDP Provider Integration

- [x] Add CDP-backed managed wallet adapter.
- [x] Add CDP account provisioning request path.
- [x] Add CDP typed-data signing request path.
- [x] Build CDP Bearer JWT auth in backend.
- [x] Keep deterministic insecure fallback for local/test mode.
- [ ] Verify `X-Wallet-Auth` production token format against live CDP credentials.
- [ ] Run live end-to-end provisioning/signing against real CDP account.

## FireFly Provider Integration

- [x] Add FireFly-backed managed wallet adapter as a selectable backend.
- [x] Generate Keystore V3 wallet files for FireFly-compatible fswallet storage.
- [x] Call FireFly Signer over JSON-RPC for managed x402 typed-data signing.
- [x] Keep deterministic insecure fallback for local/test mode.
- [x] Add admin toggles for FireFly availability, active backend, auto-pay, legacy mode, and external-wallet exposure.
- [ ] Run live end-to-end provisioning/signing against a real FireFly Signer instance.
- [ ] Verify signer auth and keystore mount hardening in the target environment.

## x402 Payment Execution

- [x] Generate x402 payment payloads automatically from managed wallet flow.
- [x] Preserve payment identifier for MCP replay detection.
- [x] Persist wallet provider/address attribution on x402 intents.
- [ ] Align every internal requirement field with canonical x402 v2 naming everywhere.
- [ ] Run live facilitator verification using real provider-signed payloads.

## Advanced External Wallets

- [ ] Add external wallet connect UX for advanced users.
- [ ] Support MetaMask / Coinbase Wallet / Rabby as optional self-custody mode.
- [ ] Keep marketplace-managed wallet as default.
- [ ] Ensure external-wallet payments still map to marketplace user identity and ledger.

## Verification

- [x] Frontend typecheck passes.
- [x] Frontend lint passes with existing warnings only.
- [x] Targeted backend managed-wallet tests pass.
- [ ] Run live production-path verification with real FireFly or CDP credentials.
