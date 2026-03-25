#!/usr/bin/env python3
"""Check a JSONL state dump for lines missing the 'address' field."""

import json
import os
import sys


def main():
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} <state_dump.jsonl>", file=sys.stderr)
        sys.exit(1)

    path = sys.argv[1]
    total_size = os.path.getsize(path)
    bytes_read = 0
    last_pct = -1

    with open(path) as f:
        for i, line in enumerate(f, start=1):
            bytes_read += len(line.encode())
            pct = int(bytes_read * 100 / total_size)
            if pct != last_pct:
                print(f"\r  {pct}% ({i:,} lines)", end="", flush=True, file=sys.stderr)
                last_pct = pct
            line = line.strip()
            if not line:
                continue
            try:
                entry = json.loads(line)
            except json.JSONDecodeError as e:
                print(f"Line {i}: INVALID JSON: {e}")
                print(f"  Content: {line[:200]}")
                continue

            if "address" not in entry:
                if "key" in entry:
                    print(f"\nLine {i}: missing 'address' field (key: {entry['key']})")
                else:
                    print(f"\nLine {i}: missing 'address' field, no key either")
                print(f"  Keys: {list(entry.keys())}")
                print(f"  Content: {line[:200]}")

    print(f"\nDone. Checked {i:,} lines.", file=sys.stderr)


if __name__ == "__main__":
    main()
