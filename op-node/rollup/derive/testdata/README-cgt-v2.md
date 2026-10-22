# CGTv2 derivation fixtures

`cgt_v2_signal_v1.json` is copied byte-for-byte from celo-kona commit `9d38b03a5271d822fcb9a7999317edad3691001b`, `crates/celo-alloy/consensus/testdata/cgt_v2_signal_v1.json`. The Go and Rust constructors must produce the same encoded transaction, source hash, and transaction hash.

Tests use a synthetic L1 bridge at `0xcc`, with `migratedAmount` in slot 0 and `escrowSeeded` in slot 1, byte offset 0. These are test inputs, not a deployed contract layout. The tests build real account/storage tries and verify their inclusion/exclusion proofs and receipt roots.

The production L1 bridge proxy address, final snapshot storage layout, final contract revision, and migration total supply remain unresolved. There is no runtime override. Until the bridge address/layout is compiled into the network rules, detecting the enable event returns an explicit configuration error. The provisional SystemConfig feature mapping slot (109) comes from the v6 compiler snapshot at Optimism commit `20bb4480868ce6071d9b4a27bc6b5557755f54b6`; it must be confirmed against the final migration artifact. The executor owns the compiled total supply and checks `S - A`.

Kona's authenticated signal reconstruction and node/prover devnet parity are separate follow-up work. These Go tests do not establish full migration readiness.
