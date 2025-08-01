#!/usr/bin/env python3
import json
import sys

def convert_to_anvil_state(input_file, output_file):
    """Convert state dump to Anvil state format with correct nonce values"""

    # Read the state dump
    with open(input_file, 'r') as f:
        state_dump = json.load(f)

    # Convert nonce values from hex strings to integers
    anvil_accounts = {}
    for address, account in state_dump.items():
        anvil_accounts[address] = {
            "balance": account["balance"],
            "code": account["code"],
            "nonce": int(account["nonce"], 16) if account["nonce"].startswith("0x") else int(account["nonce"]),
            "storage": account["storage"]
        }

    # Create Anvil state format
    anvil_state = {
        "accounts": anvil_accounts
    }

    # Write the Anvil state file
    with open(output_file, 'w') as f:
        json.dump(anvil_state, f, indent=2)

    print(f"Created Anvil state file: {output_file}")

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python convert_to_anvil_state.py <state_dump_file> <anvil_state_file>")
        sys.exit(1)

    input_file = sys.argv[1]
    output_file = sys.argv[2]
    convert_to_anvil_state(input_file, output_file)
