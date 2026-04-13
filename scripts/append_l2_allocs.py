#!/usr/bin/env python3
"""Appends Celo L2 migration allocs to a reth state dump for use with `init-state --without-ovm`.

Downloads l2-allocs.json (the accounts injected during the Cel2 migration) and appends them
to a copy of the state dump JSONL file. The original file is never modified.

The state dump format expected by reth's `init_from_state_dump` is:
  - Line 1: {"root": "0x..."}  (state root)
  - Lines 2+: {"address": "0x...", "balance": "0x0", "nonce": 1, "code": "0x...", "storage": {...}}

The Cel2 migration header's state_root already accounts for these alloc entries being present.

Usage:
  python3 append_l2_allocs.py <state_dump.jsonl> [output.jsonl]

If output path is omitted, creates <state_dump.jsonl>.with-allocs.jsonl.
The state root on line 1 is replaced with the Cel2 migration header root, since reth
verifies it against CEL2_HEADER.state_root before importing.
"""

import argparse
import json
import os
import shutil
import sys
import urllib.request

L2_ALLOCS_URL = "https://storage.googleapis.com/cel2-rollup-files/celo/l2-allocs.json"

# Cel2 migration header state root (includes both L1 state and L2 allocs).
CEL2_STATE_ROOT = "0xed980641a4bd4d2e84c6c8db980b7f05e95733c92be2e0045db3735efeb1d807"


def replace_state_root(path, new_root):
    """Replace the state root on line 1 of a JSONL state dump file in place."""
    with open(path, "rb") as f:
        first_line = f.readline()

    header = json.loads(first_line)
    old_root = header["root"]
    print(f"Replacing state root: {old_root} -> {new_root}")

    header["root"] = new_root
    new_first_line = json.dumps(header, separators=(",", ":")).encode() + b"\n"

    print(f"Old first line: {first_line}")
    print(f"New first line: {new_first_line}")
    if len(new_first_line) != len(first_line):
        print(
            f"Error: new header is {len(new_first_line)} bytes but original is "
            f"{len(first_line)} bytes. In-place replacement requires matching byte "
            f"lengths. Pad or adjust the header manually.",
            file=sys.stderr,
        )
        sys.exit(1)

    with open(path, "r+b") as f:
        f.write(new_first_line)


def main():
    parser = argparse.ArgumentParser(
        description="Append Celo L2 migration allocs to a reth state dump."
    )
    parser.add_argument("state_dump", help="Path to the state dump JSONL file")
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

    total_bytes = os.path.getsize(state_dump_path)
    copied_bytes = 0
    chunk_size = 64 * 1024 * 1024  # 64MB
    print(
        f"Copying {state_dump_path} to {output_path} ({total_bytes / (1024**3):.1f} GB)..."
    )
    with open(state_dump_path, "rb") as src, open(output_path, "wb") as dst:
        while True:
            chunk = src.read(chunk_size)
            if not chunk:
                break
            dst.write(chunk)
            copied_bytes += len(chunk)
            print(
                f"\r  {copied_bytes / (1024**3):.1f} / {total_bytes / (1024**3):.1f} GB",
                end="",
                flush=True,
            )
    shutil.copystat(state_dump_path, output_path)
    print()

    print(f"Downloading l2-allocs.json from {L2_ALLOCS_URL}...")
    with urllib.request.urlopen(L2_ALLOCS_URL) as resp:
        allocs = json.load(resp)

    print(f"Appending {len(allocs)} accounts to {output_path}...")
    with open(output_path, "a") as out:
        for address, account in allocs.items():
            entry = {"address": address}
            entry["balance"] = account.get("balance", "0x0")
            if "nonce" in account:
                entry["nonce"] = account["nonce"]
            if "code" in account:
                entry["code"] = account["code"]
            if "storage" in account:
                entry["storage"] = account["storage"]
            out.write(json.dumps(entry) + "\n")

    replace_state_root(output_path, CEL2_STATE_ROOT)

    print("Done.")


if __name__ == "__main__":
    main()
