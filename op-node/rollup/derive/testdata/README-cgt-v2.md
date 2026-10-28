# CGTv2 derivation fixtures

`cgtv2/` contains the frozen synthetic chain-1337 NUT bundle, prestate, RLP transactions and provenance manifest. The Go constructor must reproduce the shared transaction bytes, source hashes and transaction hashes, including the separate reserve-funding deposit.

`cgtv2-origin.json` contains a shared L1 header, receipt and expected L2 transactions. Tests bind the header to the requested origin hash, verify receipts against its receiptsRoot, and reconstruct deposit log indices rather than trusting RPC metadata. Every Portal deposit from the activation origin executes before the upgrade group, including deposits emitted after `EscrowSeeded`.

Activation uses the configured bridge's `EscrowSeeded(address,uint256)` event with the configured Portal as its indexed caller. Its amount determines reserve funding as fixed total supply minus migrated amount. No account/storage proof or migration-status bootstrap/cache is required. The synthetic bridge is `0xcc` and its fixed supply is 1,000; production profiles use the fixed CELO supply and separate network bundles.

Production bundle files under `upgrades/cgtv2/` are empty placeholders. Final L1 bridge addresses, reviewed network bundles and calibrated gas limits must be registered before activation can succeed. Cross-client fixture parity does not establish production node/prover/devnet readiness.
