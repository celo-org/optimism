# CGTv2 derivation fixtures

`cgtv2/` contains the frozen synthetic chain-1337 NUT bundle, prestate, RLP transactions and provenance manifest. The Go constructor must reproduce the shared transaction bytes, source hashes and transaction hashes, including the separate reserve-funding deposit.

`cgtv2-origin.json` contains a shared L1 header, receipt and expected L2 transactions. The fixture test decodes raw receipts and reconstructs log metadata before passing them to the attributes builder. Receipt-root and RPC metadata validation are covered by the existing receipt-source tests. Every Portal deposit from the activation origin executes before the upgrade group, including deposits emitted after `EscrowSeeded`.

Activation uses the configured bridge's `EscrowSeeded(address,uint256)` event with the configured Portal as its indexed caller. Its amount determines reserve funding as fixed total supply minus migrated amount. No account/storage proof or migration-status bootstrap/cache is required. The synthetic bridge is `0xcc` and its fixed supply is 1,000; production profiles use the fixed CELO supply and separate network bundles.

Production bundle files under `upgrades/cgtv2/` are empty placeholders. Final L1 bridge addresses, reviewed network bundles and calibrated gas limits must be registered before activation can succeed. Cross-client fixture parity does not establish production node/prover/devnet readiness.

`cgtv2-receipt-scan.json` holds receipt lists with the expected result of the event scan on each: no trigger, escrow, retry, or a malformed-event error. The scan test runs these cases against `cgtV2EscrowAmount` and asserts the file's keccak256. celo-kona keeps an identical copy at `crates/kona/derive/testdata/`, and its test asserts the same hash, so a one-sided edit fails a test. Nothing compares the two repos automatically: a change must update both files and both hash constants.
