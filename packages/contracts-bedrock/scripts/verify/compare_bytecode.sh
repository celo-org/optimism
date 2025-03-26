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

# Now compare ignoring immutables
if [ "$DEPLOYED_BYTECODE" = "$LOCAL_BYTECODE" ]; then
  echo "$ARTIFACT_FILE Success: Deployed bytecode matches local artifact (excluding immutables)."
else
  echo "$ARTIFACT_FILE Mismatch: Bytecode differs beyond immutables."
  echo "Deployed: $DEPLOYED_BYTECODE"
  echo "Local:    $LOCAL_BYTECODE"
fi
