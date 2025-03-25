#!/bin/bash
# Usage: ./compare_bytecode.sh <contract-address> <artifact-file>
#
# This script fetches the deployed bytecode using Foundry's "cast code"
# and compares it against the "deployedBytecode" field from your local artifact.
# Ensure that "jq" is installed to parse the JSON artifact.
#
# Optionally, you can modify the script to include an RPC URL by adding the
# "--rpc-url <your_rpc_url>" option to the cast command.

# Check if both arguments are provided
if [ "$#" -lt 2 ]; then
  echo "Usage: $0 <contract-address> <artifact-file>"
  exit 1
fi

CONTRACT_ADDRESS=$1
ARTIFACT_FILE=$2

# Retrieve deployed bytecode from chain using cast
echo "Fetching deployed bytecode for contract $CONTRACT_ADDRESS..."
DEPLOYED_BYTECODE=$(cast code "$CONTRACT_ADDRESS" --rpc-url https://eth.llamarpc.com | tr -d '\n')

if [ -z "$DEPLOYED_BYTECODE" ]; then
  echo "Error: Could not fetch bytecode. Please check the contract address or your network configuration."
  exit 1
fi

# Extract the deployed bytecode from the artifact file using jq
# This assumes your artifact JSON contains a key "deployedBytecode"
echo "Extracting local deployedBytecode from artifact $ARTIFACT_FILE..."
LOCAL_BYTECODE=$(jq -r '.deployedBytecode.object' "$ARTIFACT_FILE" | tr -d '\n')

if [ -z "$LOCAL_BYTECODE" ]; then
  echo "Error: Could not extract deployedBytecode from the artifact. Verify the artifact file format."
  exit 1
fi

# Compare the two bytecodes
if [ "$DEPLOYED_BYTECODE" = "$LOCAL_BYTECODE" ]; then
  echo "Success: The deployed bytecode matches the local artifact."
else
  echo "Mismatch: The deployed bytecode does not match the local artifact."
  # Optionally, you can output both for manual diff
  echo "Deployed bytecode: $DEPLOYED_BYTECODE"
  echo "Local artifact bytecode: $LOCAL_BYTECODE"
fi
