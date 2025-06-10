#!/bin/bash
# Usage: ./compare_bytecode_ignore_immutables.sh <contract-address> <artifact-file>

if [ "$#" -lt 2 ]; then
  echo "Usage: $0 <contract-address> <artifact-file>"
  exit 1
fi

CONTRACT_ADDRESS=$1
ARTIFACT_FILE=$2

# Fetch deployed bytecode from chain
DEPLOYED_BYTECODE=$(cast code "$CONTRACT_ADDRESS" --rpc-url https://eth.llamarpc.com | tr -d '\n')

if [ -z "$DEPLOYED_BYTECODE" ]; then
  echo "Error: Failed to fetch bytecode."
  exit 1
fi

# Get local bytecode
LOCAL_BYTECODE=$(jq -r '.deployedBytecode.object' "$ARTIFACT_FILE" | tr -d '\n')

if [ -z "$LOCAL_BYTECODE" ]; then
  echo "Error: Failed to extract local bytecode."
  exit 1
fi

# Special exception for SuperchainConfig version diff
if grep -q "SuperchainConfig" "$ARTIFACT_FILE"; then
  # Replace metadata version "1.1.1-beta.1" with "1.1.0" since SuperchainConfig was deployed with version "1.1.0" instead of "1.1.1-beta.1" (no other changes)
  LOCAL_BYTECODE=$(echo "$LOCAL_BYTECODE" | sed 's/600c81526020017f312e312e312d626574612e31/600581526020017f312e312e3000000000000000/g')
fi

# Replace immutables with 0000
IMMUTABLES=$(jq -c '.deployedBytecode.immutableReferences' "$ARTIFACT_FILE")

replace_with_zeros() {
  local BYTECODE=$1
  local START=$2
  local LENGTH=$3
  local PREFIX=${BYTECODE:0:START}
  local SUFFIX=${BYTECODE:START+LENGTH}
  local ZEROS=$(printf '%*s' "$LENGTH" '' | tr ' ' '0')
  echo "$PREFIX$ZEROS$SUFFIX"
}

if [ "$IMMUTABLES" != "null" ]; then
  for entry in $(echo "$IMMUTABLES" | jq -c '.[] | .[]'); do
    START=$(($(echo "$entry" | jq '.start * 2')))
    LENGTH=$(($(echo "$entry" | jq '.length * 2')))

    DEPLOYED_BYTECODE=$(replace_with_zeros "$DEPLOYED_BYTECODE" "$((START + 2))" "$LENGTH")
  done
fi

# Now compare ignoring immutables and version diff
if [ "$DEPLOYED_BYTECODE" = "$LOCAL_BYTECODE" ]; then
  echo "$ARTIFACT_FILE Success: Deployed bytecode matches local artifact (excluding immutables/version diff)."
else
  echo "$ARTIFACT_FILE Mismatch: Bytecode differs beyond immutables/version diff."
  echo "Deployed: $DEPLOYED_BYTECODE"
  echo "Local:    $LOCAL_BYTECODE"
fi
