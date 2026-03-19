#!/usr/bin/env python3
"""Merge Celo L1 state dump with L2 migration allocs for reth import.

Downloads l2-allocs.json and merges it with the L1 state dump, applying the
same account merge logic as the Go migration. Also fixes the zero address dump
bug and sets the treasury balance.

Merge rules (from celo-org/optimism@celo-rebase-12 state.go applyAllocsToState):
  1. Alloc account not in L1       -> create with alloc data, balance=0
  2. In L1, allowlisted (false)    -> skip alloc, keep L1 account as-is
  3. In L1, allowlisted (true)     -> overwrite code only from alloc
  4. In L1, not allowlisted, has code/nonce -> error
  5. In L1, not allowlisted, balance-only   -> take alloc code/nonce/storage, keep L1 balance

Treasury balance is set to the value from migration block 31,056,500
(= 1e27 - celoTokenTotalSupply + treasuryL1Balance).

Usage:
  python3 append_l2_allocs.py <state_dump.jsonl> [output.jsonl]

If output path is omitted, creates <state_dump>.with-allocs.jsonl.
"""

import argparse
import json
import os
import re
import sys
import urllib.request

L2_ALLOCS_URL = "https://storage.googleapis.com/cel2-rollup-files/celo/l2-allocs.json"

# Cel2 migration header state root (includes both L1 state and L2 allocs).
CEL2_STATE_ROOT = "0xed980641a4bd4d2e84c6c8db980b7f05e95733c92be2e0045db3735efeb1d807"

# CeloToken contract address (mainnet).
CELO_TOKEN = "0x471ece3750da237f93b8e339c536989b8978a438"

# Treasury address (mainnet).
TREASURY = "0x7a8c7a833565fc428cdfba20fe03fafb178a434f"

# Treasury balance at migration block 31,056,500, queried from L2 archive node.
# Equals: TREASURY_CEILING - celoTokenTotalSupply + treasuryL1Balance
TREASURY_BALANCE = 292574923875528830166719349

# Ceiling used in setupUnreleasedTreasury (1 billion CELO in wei).
TREASURY_CEILING = 1_000_000_000 * 10**18

# CeloToken totalSupply storage slot (used for verification only).
TOTAL_SUPPLY_SLOT = "0x0000000000000000000000000000000000000000000000000000000000000002"

# L1 dump bug: zero address has "key" (keccak hash) instead of "address".
ZERO_ADDR_KEY = '"key":"0x5380c7b7ae81a58eb98d9c78de4a1fd7fd9535fc953ed2be602daaa41767312a"'
ZERO_ADDR_FIX = '"address":"0x0000000000000000000000000000000000000000"'

# Accounts where L2 allocs should not blindly overwrite the L1 state.
# False = skip alloc entirely, True = overwrite code only.
ALLOWLIST = {
    "0x000000000022d473030f116ddee9f6b43ac78ba3": False,  # Permit2
    "0x0000000071727de22e5e9d8baf0edac6f37da032": False,  # EntryPoint_v070
    "0x13b0d85ccb8bf860b6b79af3029fca081ae9bef2": True,   # Create2Deployer
    "0x4e59b44847b379578588920ca78fbf26c0b4956c": False,  # DeterministicDeploymentProxy
    "0x5ff137d4b0fdcd49dca30c7cf57e578a026d2789": False,  # EntryPoint_v060
    "0x69f4d1788e39c87893c980c06edf4b7f686e2938": False,  # Safe_v130
    "0x7fc98430eaedbb6070b35b39d798725049088348": False,  # SenderCreator_v060
    "0x914d7fec6aac8cd542e72bca78b30650d45643d7": False,  # SafeSingletonFactory
    "0x998739bfdaadde7c933b942a68053933098f9eda": False,  # MultiSend_v130
    "0xa1dabef33b3b82c7814b6d82a79e50f4ac44102b": False,  # MultiSendCallOnly_v130
    "0xba5ed099633d3b313e4d5f7bdc1305d3c28ba5ed": False,  # CreateX
    "0xca11bde05977b3631167028862be2a173976ca11": False,  # Multicall3
    "0xefc2c1444ebcc4db75e7613d20c6a62ff67a167c": False,  # SenderCreator_v070
    "0xfb1bffc9d739b8d520daf37df666da4c687191ea": False,  # SafeL2_v130
}

_ADDRESS_RE = re.compile(r'"address"\s*:\s*"(0x[0-9a-fA-F]{40})"')


def extract_address(line):
    """Extract address from a dump line via regex. Returns lowercase or None."""
    m = _ADDRESS_RE.search(line)
    return m.group(1).lower() if m else None


def parse_balance(s):
    """Parse a balance from decimal or hex string."""
    if s.startswith(("0x", "0X")):
        return int(s, 16)
    return int(s)


def account_line(entry):
    """Serialize an account dict to a compact JSONL line."""
    return json.dumps(entry, separators=(",", ":")) + "\n"


def build_entry(address, alloc):
    """Build an account entry from alloc data with zero balance."""
    entry = {"address": address, "balance": "0x0"}
    if "nonce" in alloc:
        entry["nonce"] = alloc["nonce"]
    if "code" in alloc:
        entry["code"] = alloc["code"]
    if "storage" in alloc:
        entry["storage"] = alloc["storage"]
    return entry


def merge_alloc(l1_account, alloc, address):
    """Merge an L2 alloc into an existing L1 account per Go migration rules.

    Returns the merged account dict. Raises ValueError for case 4
    (existing account with code or nonzero nonce not in allowlist).
    """
    if address in ALLOWLIST:
        if not ALLOWLIST[address]:
            # Case 2: keep L1 account unchanged
            print(f"  Merge {address}: case 2 (allowlist skip, keep L1)")
            return dict(l1_account)
        # Case 3: overwrite code only
        print(f"  Merge {address}: case 3 (allowlist, overwrite code only)")
        merged = dict(l1_account)
        if "code" in alloc:
            merged["code"] = alloc["code"]
        return merged

    # Case 4 check: L1 account has code or nonce > 0
    l1_code = l1_account.get("code", "0x")
    l1_nonce = l1_account.get("nonce", 0)
    if isinstance(l1_nonce, str):
        l1_nonce = int(l1_nonce, 0) if l1_nonce.startswith(("0x", "0X")) else int(l1_nonce)
    if l1_code not in ("0x", "") or l1_nonce > 0:
        raise ValueError(
            f"L1 account {address} has code={l1_code!r} nonce={l1_nonce} "
            f"but is not in the allowlist"
        )

    # Case 5: balance-only L1 account — take alloc data, keep L1 balance
    l1_balance = l1_account.get("balance", "0x0")
    print(f"  Merge {address}: case 5 (balance-only L1, balance={l1_balance})")
    entry = build_entry(address, alloc)
    entry["balance"] = l1_balance
    return entry


def main():
    parser = argparse.ArgumentParser(
        description="Merge Celo L1 state dump with L2 migration allocs for reth import."
    )
    parser.add_argument("state_dump", help="Path to the L1 state dump JSONL file")
    parser.add_argument(
        "output",
        nargs="?",
        help="Output path (default: <state_dump>.with-allocs.jsonl)",
    )
    args = parser.parse_args()

    state_dump_path = args.state_dump
    output_path = args.output or state_dump_path + ".with-allocs.jsonl"

    if not os.path.exists(state_dump_path):
        print(f"Error: {state_dump_path} does not exist", file=sys.stderr)
        sys.exit(1)

    # Phase 1: Load L2 allocs
    print(f"Downloading l2-allocs.json from {L2_ALLOCS_URL}...")
    with urllib.request.urlopen(L2_ALLOCS_URL) as resp:
        raw_allocs = json.load(resp)
    allocs = {addr.lower(): data for addr, data in raw_allocs.items()}
    handled = set()
    print(f"Loaded {len(allocs)} L2 alloc accounts.")

    # State tracked during L1 dump pass
    total_supply = None
    treasury_l1_account = None
    treasury_l1_balance = 0
    line_count = 0
    merged_count = 0

    # Phase 2: Stream L1 dump with inline modifications
    print(f"Processing {state_dump_path} -> {output_path}...")
    with open(state_dump_path, "r") as inp, open(output_path, "w") as out:
        for line_num, line in enumerate(inp):
            # Line 1: replace state root
            if line_num == 0:
                header = json.loads(line)
                old_root = header["root"]
                header["root"] = CEL2_STATE_ROOT
                out.write(json.dumps(header, separators=(",", ":")) + "\n")
                print(f"State root: {old_root} -> {CEL2_STATE_ROOT}")
                continue

            # Zero address fix: replace key hash with address field
            if ZERO_ADDR_KEY in line:
                line = line.replace(ZERO_ADDR_KEY, ZERO_ADDR_FIX)
                print("Fixed zero address entry (key -> address)")

            address = extract_address(line)
            if address is None:
                out.write(line)
                continue

            line_count += 1
            if line_count % 5_000_000 == 0:
                print(f"  {line_count} accounts processed...")

            # CeloToken: extract totalSupply for verification
            if address == CELO_TOKEN:
                account = json.loads(line)
                storage = account.get("storage", {})
                for slot in (TOTAL_SUPPLY_SLOT, "0x2"):
                    if slot in storage:
                        total_supply = int(storage[slot], 16)
                        print(f"CeloToken totalSupply: {total_supply}")
                        break
                if address in allocs:
                    merged = merge_alloc(account, allocs[address], address)
                    out.write(account_line(merged))
                    handled.add(address)
                    merged_count += 1
                else:
                    out.write(line)
                continue

            # Treasury: save for phase 4, don't write yet
            if address == TREASURY:
                treasury_l1_account = json.loads(line)
                balance_str = treasury_l1_account.get("balance", "0x0")
                treasury_l1_balance = parse_balance(balance_str)
                print(f"Treasury L1 balance: {treasury_l1_balance}")
                continue

            # L2 alloc merge
            if address in allocs:
                l1_account = json.loads(line)
                merged = merge_alloc(l1_account, allocs[address], address)
                out.write(account_line(merged))
                handled.add(address)
                merged_count += 1
                continue

            # Default: pass through unchanged
            out.write(line)

        # Phase 3: Append remaining allocs (new accounts not in L1)
        new_count = 0
        for address, alloc in allocs.items():
            if address in handled or address == TREASURY:
                continue
            out.write(account_line(build_entry(address, alloc)))
            new_count += 1

        # Phase 4: Write treasury with hardcoded balance
        treasury_alloc = allocs.get(TREASURY)
        if treasury_alloc and treasury_l1_account:
            print(f"Treasury {TREASURY}: L1 + alloc, merge then set balance")
            treasury_entry = merge_alloc(
                treasury_l1_account, treasury_alloc, TREASURY
            )
            treasury_entry["balance"] = hex(TREASURY_BALANCE)
        elif treasury_alloc:
            print(f"Treasury {TREASURY}: alloc only, set balance")
            treasury_entry = build_entry(TREASURY, treasury_alloc)
            treasury_entry["balance"] = hex(TREASURY_BALANCE)
        elif treasury_l1_account:
            print(f"Treasury {TREASURY}: L1 only, set balance")
            treasury_entry = dict(treasury_l1_account)
            treasury_entry["balance"] = hex(TREASURY_BALANCE)
        else:
            print(f"Treasury {TREASURY}: not in L1 or allocs, creating")
            treasury_entry = {"address": TREASURY, "balance": hex(TREASURY_BALANCE)}
        out.write(account_line(treasury_entry))

    print(f"\nProcessed {line_count} L1 accounts")
    print(f"Merged {merged_count} existing accounts with L2 allocs")
    print(f"Appended {new_count} new accounts from L2 allocs")
    print(f"Treasury balance: {TREASURY_BALANCE} ({hex(TREASURY_BALANCE)})")

    # Phase 5: Verify treasury balance
    if total_supply is not None:
        expected = TREASURY_CEILING - total_supply + treasury_l1_balance
        if expected != TREASURY_BALANCE:
            print(
                f"\nERROR: Treasury balance mismatch!\n"
                f"  ceiling - totalSupply + l1Balance = "
                f"{TREASURY_CEILING} - {total_supply} + {treasury_l1_balance} = {expected}\n"
                f"  Hardcoded: {TREASURY_BALANCE}",
                file=sys.stderr,
            )
            sys.exit(1)
        print(
            f"Treasury balance verified: "
            f"{TREASURY_CEILING} - {total_supply} + {treasury_l1_balance} = {TREASURY_BALANCE}"
        )
    else:
        print(
            "Warning: CeloToken totalSupply not found, skipping treasury verification"
        )

    print("\nDone.")


if __name__ == "__main__":
    main()
