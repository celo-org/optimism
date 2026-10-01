#!/bin/bash
# Enclave Batcher Runner Script
# Supports both local (docker-compose) and AWS ECS deployments

set -e

# Required environment variables - will fail if not set
: "${L1_RPC_URL:?Error: L1_RPC_URL is required}"
: "${L2_RPC_URL:?Error: L2_RPC_URL is required}"
: "${ROLLUP_RPC_URL:?Error: ROLLUP_RPC_URL is required}"
: "${ESPRESSO_URL1:?Error: ESPRESSO_URL1 is required}"
: "${OPERATOR_PRIVATE_KEY:?Error: OPERATOR_PRIVATE_KEY is required}"
: "${ESPRESSO_ATTESTATION_SERVICE_URL:?Error: ESPRESSO_ATTESTATION_SERVICE_URL is required}"
: "${EIGENDA_PROXY_URL:?Error: EIGENDA_PROXY_URL is required}"

# Optional configuration with defaults
TAG="${TAG:-op-batcher-enclavetool}"
ESPRESSO_URL2="${ESPRESSO_URL2:-$ESPRESSO_URL1}"  # Default to same as URL1 if not set
ESPRESSO_ORIGIN_HEIGHT_ESPRESSO="${ESPRESSO_ORIGIN_HEIGHT_ESPRESSO:-0}"
ESPRESSO_ORIGIN_HEIGHT_L2="${ESPRESSO_ORIGIN_HEIGHT_L2:-0}"
ENCLAVE_DEBUG="${ENCLAVE_DEBUG:-false}"
MONITOR_INTERVAL="${MONITOR_INTERVAL:-30}"
MEMORY_MB="${ENCLAVE_MEMORY_MB:-4096}"
CPU_COUNT="${ENCLAVE_CPU_COUNT:-2}"
MAX_CHANNEL_DURATION="${MAX_CHANNEL_DURATION:-2}"
TARGET_NUM_FRAMES="${TARGET_NUM_FRAMES:-1}"
MAX_L1_TX_SIZE_BYTES="${MAX_L1_TX_SIZE_BYTES:-120000}"
ALTDA_MAX_CONCURRENT_DA_REQUESTS="${ALTDA_MAX_CONCURRENT_DA_REQUESTS:-1}"

# Deployment mode detection
DEPLOYMENT_MODE="${DEPLOYMENT_MODE:-aws}"  # 'local' or 'aws'

# Get batch authenticator address from env var or deployment state
if [ -n "$BATCH_AUTHENTICATOR_ADDRESS" ]; then
    echo "Using BATCH_AUTHENTICATOR_ADDRESS from environment variable"
else
    address_from_state=$(jq -r '.opChainDeployments[0].batchAuthenticatorAddress' /app/deployment/deployer/state.json 2>/dev/null)
    if [ -n "$address_from_state" ] && [ "$address_from_state" != "null" ]; then
        BATCH_AUTHENTICATOR_ADDRESS="$address_from_state"
        echo "Using BATCH_AUTHENTICATOR_ADDRESS from state.json"
    else
        echo "WARNING: BATCH_AUTHENTICATOR_ADDRESS not found in environment or state.json"
        BATCH_AUTHENTICATOR_ADDRESS=""
    fi
fi

export BATCH_AUTHENTICATOR_ADDRESS

# Get light client address from env var or use default
if [ -n "$ESPRESSO_LIGHT_CLIENT_ADDR" ]; then
    echo "Using ESPRESSO_LIGHT_CLIENT_ADDR from environment variable"
else
    # Decaf light client address for ETH Sepolia
    ESPRESSO_LIGHT_CLIENT_ADDR="0x303872bb82a191771321d4828888920100d0b3e4"
    echo "ESPRESSO_LIGHT_CLIENT_ADDR not set, using default"
fi

# Override OP_BATCHER_ESPRESSO_LIGHT_CLIENT_ADDR so the batcher's env var matches,
# preventing any outer deployment env from leaking a stale value into the enclave.
export OP_BATCHER_ESPRESSO_LIGHT_CLIENT_ADDR="$ESPRESSO_LIGHT_CLIENT_ADDR"

echo "=== Enclave Batcher Configuration ==="
echo "Deployment Mode: $DEPLOYMENT_MODE"
echo "L1 RPC URL: $L1_RPC_URL"
echo "L2 RPC URL: $L2_RPC_URL"
echo "Rollup RPC URL: $ROLLUP_RPC_URL"
echo "Espresso URLs: $ESPRESSO_URL1, $ESPRESSO_URL2"
echo "Attestation service url: $ESPRESSO_ATTESTATION_SERVICE_URL"
echo "EigenDA Proxy URL: $EIGENDA_PROXY_URL"
echo "Batch Authenticator Address: ${BATCH_AUTHENTICATOR_ADDRESS:-[not set]}"
echo "Light Client Address: $ESPRESSO_LIGHT_CLIENT_ADDR"
echo "Espresso Origin Height: $ESPRESSO_ORIGIN_HEIGHT_ESPRESSO"
echo "L2 Origin Height: $ESPRESSO_ORIGIN_HEIGHT_L2"
echo "Debug Mode: $ENCLAVE_DEBUG"
echo "Monitor Interval: $MONITOR_INTERVAL seconds"
echo "Memory: ${MEMORY_MB}MB"
echo "CPU Count: $CPU_COUNT"
echo "KMS Signer Endpoint: ${SIGNER_ENDPOINT:-[not set]}"
echo "KMS Signer Address: ${SIGNER_ADDRESS:-[not set]}"
echo "Max Channel Duration: $MAX_CHANNEL_DURATION"
echo "Target Num Frames: $TARGET_NUM_FRAMES"
echo "Max L1 Tx Size Bytes: $MAX_L1_TX_SIZE_BYTES"
echo "AltDA Max Concurrent DA Requests: $ALTDA_MAX_CONCURRENT_DA_REQUESTS"
echo "====================================="

# Batcher arguments
BATCHER_ARGS="--l1-eth-rpc=$L1_RPC_URL"
BATCHER_ARGS="$BATCHER_ARGS,--l2-eth-rpc=$L2_RPC_URL"
BATCHER_ARGS="$BATCHER_ARGS,--rollup-rpc=$ROLLUP_RPC_URL"
BATCHER_ARGS="$BATCHER_ARGS,--espresso.enabled=true"
BATCHER_ARGS="$BATCHER_ARGS,--espresso.urls=$ESPRESSO_URL1"
BATCHER_ARGS="$BATCHER_ARGS,--espresso.urls=$ESPRESSO_URL2"
BATCHER_ARGS="$BATCHER_ARGS,--espresso.espresso-attestation-service=$ESPRESSO_ATTESTATION_SERVICE_URL"
BATCHER_ARGS="$BATCHER_ARGS,--espresso.origin-height-espresso=$ESPRESSO_ORIGIN_HEIGHT_ESPRESSO"
BATCHER_ARGS="$BATCHER_ARGS,--espresso.origin-height-l2=$ESPRESSO_ORIGIN_HEIGHT_L2"

# Use KMS signer if endpoint+address are provided, otherwise fall back to private key or test mnemonic.
if [ -n "${SIGNER_ENDPOINT:-}" ] || [ -n "${SIGNER_ADDRESS:-}" ]; then
    if [ -n "${SIGNER_ENDPOINT:-}" ] && [ -n "${SIGNER_ADDRESS:-}" ]; then
        echo "Using KMS signer at $SIGNER_ENDPOINT (address: $SIGNER_ADDRESS)"
        BATCHER_ARGS="$BATCHER_ARGS,--signer.endpoint=$SIGNER_ENDPOINT"
        BATCHER_ARGS="$BATCHER_ARGS,--signer.address=$SIGNER_ADDRESS"
    else
        echo "ERROR: Both SIGNER_ENDPOINT and SIGNER_ADDRESS must be set to use KMS signer"
        exit 1
    fi
elif [ -n "$OP_BATCHER_PRIVATE_KEY" ]; then
    echo "Using OP_BATCHER_PRIVATE_KEY for authentication"
    BATCHER_ARGS="$BATCHER_ARGS,--private-key=$OP_BATCHER_PRIVATE_KEY"
else
    echo "Using test mnemonic for authentication (local development mode)"
    BATCHER_ARGS="$BATCHER_ARGS,--mnemonic=test test test test test test test test test test test junk"
    BATCHER_ARGS="$BATCHER_ARGS,--hd-path=m/44'/60'/0'/0/6"
fi

BATCHER_ARGS="$BATCHER_ARGS,--throttle.unsafe-da-bytes-lower-threshold=0"
BATCHER_ARGS="$BATCHER_ARGS,--max-channel-duration=$MAX_CHANNEL_DURATION"
BATCHER_ARGS="$BATCHER_ARGS,--target-num-frames=$TARGET_NUM_FRAMES"
BATCHER_ARGS="$BATCHER_ARGS,--max-l1-tx-size-bytes=$MAX_L1_TX_SIZE_BYTES"
BATCHER_ARGS="$BATCHER_ARGS,--max-pending-tx=32"
BATCHER_ARGS="$BATCHER_ARGS,--espresso.light-client-addr=$ESPRESSO_LIGHT_CLIENT_ADDR"
BATCHER_ARGS="$BATCHER_ARGS,--espresso.espresso-attestation-service=$ESPRESSO_ATTESTATION_SERVICE_URL"
BATCHER_ARGS="$BATCHER_ARGS,--altda.enabled=true"
BATCHER_ARGS="$BATCHER_ARGS,--altda.da-server=$EIGENDA_PROXY_URL"
BATCHER_ARGS="$BATCHER_ARGS,--altda.da-service=true"
BATCHER_ARGS="$BATCHER_ARGS,--altda.verify-on-read=false"
BATCHER_ARGS="$BATCHER_ARGS,--altda.max-concurrent-da-requests=$ALTDA_MAX_CONCURRENT_DA_REQUESTS"
BATCHER_ARGS="$BATCHER_ARGS,--altda.put-timeout=30s"
BATCHER_ARGS="$BATCHER_ARGS,--altda.get-timeout=30s"
BATCHER_ARGS="$BATCHER_ARGS,--data-availability-type=calldata"

# Add debug arguments if enabled
if [ "$ENCLAVE_DEBUG" = "true" ]; then
    BATCHER_ARGS="$BATCHER_ARGS,--log.level=debug"
    echo "Debug logging enabled"
fi

if [ -n "${SIGNER_TLS_ENABLED:-}" ]; then
    BATCHER_ARGS="$BATCHER_ARGS,--signer.tls.enabled=$SIGNER_TLS_ENABLED"
fi

# Remove any stale enclave containers from previous runs.
if stale=$(docker ps -q --filter "name=batcher-enclaver") && [ -n "$stale" ]; then
    echo "Removing stale enclave containers: $stale"
    docker rm -f "$stale"
fi

# Build the EIF from the pre-built app image (built during `docker compose build`).
# This skips the expensive Docker rebuild-from-source step and goes straight to
# Enclaver EIF conversion, saving ~3-10 minutes per startup.
APP_IMAGE="${APP_IMAGE:-op-batcher-enclave-app:espresso}"
echo "Building EIF from pre-built app image: $APP_IMAGE (tag: $TAG)"

if ! enclave-tools build-eif --app-image "$APP_IMAGE" --eif-tag "$TAG" --cpu-count "$CPU_COUNT" --memory-mb "$MEMORY_MB" 2>&1 | tee /tmp/build_output.log; then
    echo "ERROR: Failed to build EIF image"
    echo "Build output was:"
    cat /tmp/build_output.log
    exit 1
fi

echo "EIF build completed successfully"

# Extract PCR0 from build output
# Works whether the line is `... PCR0: 0xABCD ...` or `... PCR0=abcd123 ...`
PCR0="$(grep -m1 -oE 'PCR0[=:][[:space:]]*(0x)?[[:xdigit:]]{64,}' /tmp/build_output.log \
       | sed -E 's/^PCR0[=:][[:space:]]*(0x)?//')"

# Register PCR0 if all required values are present
if [ -n "$PCR0" ] && [ -n "$BATCH_AUTHENTICATOR_ADDRESS" ] && [ -n "$OPERATOR_PRIVATE_KEY" ]; then
    echo "Checking if PCR0 is already registered..."

    if enclave-tools is-registered \
        --authenticator "$BATCH_AUTHENTICATOR_ADDRESS" \
        --l1-url "$L1_RPC_URL" \
        --pcr0 "$PCR0" >/dev/null 2>&1; then
        echo "PCR0 already registered: $PCR0"
        echo "Skipping registration..."
    else
        echo "PCR0 not registered. Registering PCR0: $PCR0 with authenticator: $BATCH_AUTHENTICATOR_ADDRESS"
        if ! enclave-tools register \
            --authenticator "$BATCH_AUTHENTICATOR_ADDRESS" \
            --l1-url "$L1_RPC_URL" \
            --private-key "$OPERATOR_PRIVATE_KEY" \
            --pcr0 "$PCR0"; then
            echo "ERROR: Failed to register PCR0. Cannot continue without valid registration."
            exit 1
        fi
        echo "PCR0 registration successful"
    fi
else
    echo "Skipping PCR0 registration - missing required values:"
    echo "  PCR0: ${PCR0:-[missing]}"
    echo "  BATCH_AUTHENTICATOR_ADDRESS: ${BATCH_AUTHENTICATOR_ADDRESS:-[missing]}"
    echo "  OPERATOR_PRIVATE_KEY: ${OPERATOR_PRIVATE_KEY:+[set]}"
fi

# Setup tracking files for local deployment
if [ "$DEPLOYMENT_MODE" = "local" ]; then
    PID_FILE="/tmp/enclave-tools.pid"
    CONTAINER_TRACKER_FILE="/tmp/enclave-containers.txt"
    STATUS_FILE="/tmp/enclave-status.json"

    # Cleanup function for local deployment
    cleanup() {
        echo "Cleaning up enclave resources..."
        if [ -f "$PID_FILE" ]; then
            STORED_PID=$(cat "$PID_FILE")
            if kill -0 "$STORED_PID" 2>/dev/null; then
                echo "Terminating enclave-tools process (PID: $STORED_PID)"
                kill -TERM "$STORED_PID" 2>/dev/null || true
            fi
            rm -f "$PID_FILE"
        fi

        # Force-kill tracked enclave runner containers (tracker stores
        # container IDs).  docker rm -f sends SIGKILL immediately, which
        # also causes enclaver to terminate the Nitro Enclave VM.
        if [ -f "$CONTAINER_TRACKER_FILE" ]; then
            while IFS= read -r cid; do
                if [ -n "$cid" ]; then
                    echo "Force-removing enclave container: $cid"
                    docker rm -f "$cid" 2>/dev/null || true
                fi
            done < "$CONTAINER_TRACKER_FILE"
            rm -f "$CONTAINER_TRACKER_FILE"
        fi

        rm -f "$STATUS_FILE"
        exit 0
    }

    # Setup signal handlers for local deployment
    trap cleanup SIGTERM SIGINT EXIT

    # Get Docker network for local deployment
    DOCKER_NETWORK=$(docker network ls --filter name=espresso --format "{{.Name}}" | head -1)
    if [ -z "$DOCKER_NETWORK" ]; then
        DOCKER_NETWORK="espresso_default"
    fi
    echo "Using Docker network: $DOCKER_NETWORK"
    export DOCKER_DEFAULT_NETWORK="$DOCKER_NETWORK"
    export ENCLAVE_DOCKER_NETWORK="$DOCKER_NETWORK"
fi

# Run the enclave
echo "Starting enclave with image: $TAG (args contain sensitive data and are not logged)"

enclave-tools run --image "$TAG" --args "$BATCHER_ARGS" &
ENCLAVE_TOOLS_PID=$!

# Start log capture early so we see crash output before --rm wipes the container.
for _ in $(seq 1 20); do
    EARLY_CID=$(docker ps -aq --filter "name=batcher-enclaver-" | head -1)
    [ -n "$EARLY_CID" ] && break
    sleep 0.5
done
if [ -n "$EARLY_CID" ]; then
    echo "Early log capture attached to container $EARLY_CID"
    docker logs -f "$EARLY_CID" 2>&1 | sed 's/^/[ENCLAVE-EARLY] /' | tee /tmp/enclave-early.log &
else
    echo "WARNING: enclave container did not appear within 10s for early log capture"
fi

if [ "$DEPLOYMENT_MODE" = "local" ]; then
    echo "$ENCLAVE_TOOLS_PID" > "$PID_FILE"
    echo "Enclave-tools started with PID: $ENCLAVE_TOOLS_PID (stored in $PID_FILE)"
else
    echo "Enclave-tools started with PID: $ENCLAVE_TOOLS_PID"
fi

# Wait for enclave-tools to finish starting the enclave container
echo "Waiting for enclave-tools to complete startup..."
wait $ENCLAVE_TOOLS_PID
ENCLAVE_TOOLS_EXIT_CODE=$?
echo "Enclave-tools process completed with exit code: $ENCLAVE_TOOLS_EXIT_CODE"

if [ "$DEPLOYMENT_MODE" = "local" ]; then
    rm -f "$PID_FILE"
fi

# Check if enclave-tools failed
if [ $ENCLAVE_TOOLS_EXIT_CODE -ne 0 ]; then
    echo "ERROR: enclave-tools failed with exit code $ENCLAVE_TOOLS_EXIT_CODE"
    exit $ENCLAVE_TOOLS_EXIT_CODE
fi

# Wait for container to fully initialize
sleep 5

# Find the enclave container that was started
echo "Looking for running enclave container..."
CONTAINER_NAME=$(docker ps --format "table {{.Names}}" | grep "batcher-enclaver-" | head -1)

if [ -z "$CONTAINER_NAME" ]; then
    echo "ERROR: No enclave container found after waiting."
    echo "Checking all Docker containers:"
    docker ps -a
    exit 1
fi

echo "Found enclave container: $CONTAINER_NAME"

# Get container details
CONTAINER_ID=$(docker ps --filter "name=$CONTAINER_NAME" --format "{{.ID}}" | head -1)
CONTAINER_IMAGE=$(docker inspect "$CONTAINER_NAME" --format '{{.Config.Image}}' 2>/dev/null)
STARTED_AT=$(docker inspect "$CONTAINER_NAME" --format '{{.State.StartedAt}}' 2>/dev/null)

echo "Container Details:"
echo "  ID: $CONTAINER_ID"
echo "  Image: $CONTAINER_IMAGE"
echo "  Started: $STARTED_AT"

# Setup status tracking for local deployment
if [ "$DEPLOYMENT_MODE" = "local" ]; then
    echo "$CONTAINER_ID" >> "$CONTAINER_TRACKER_FILE"

    # Create initial status file
    cat > "$STATUS_FILE" <<EOF
{
  "container_id": "$CONTAINER_ID",
  "container_name": "$CONTAINER_NAME",
  "container_image": "$CONTAINER_IMAGE",
  "started_at": "$STARTED_AT",
  "last_updated": "$(date -Iseconds)",
  "status": "running",
  "enclave_tools_exit_code": $ENCLAVE_TOOLS_EXIT_CODE
}
EOF
fi

# Start capturing container logs in background
echo "Starting log capture for container $CONTAINER_NAME"
(
    docker logs -f "$CONTAINER_NAME" 2>&1 | while read -r line; do
        echo "[ENCLAVE] $line"
    done
) &
LOG_PID=$!
echo "Log capture started with PID: $LOG_PID"

# Monitor the container
echo "Monitoring enclave container $CONTAINER_NAME..."
MONITOR_COUNT=0

while true; do
    # Check if the container is still running
    CONTAINER_STATUS=$(docker inspect "$CONTAINER_NAME" 2>/dev/null | jq -r '.[0].State.Status' 2>/dev/null || echo "")

    if [ -z "$CONTAINER_STATUS" ] || [ "$CONTAINER_STATUS" != "running" ]; then
        echo "$(date): Container $CONTAINER_NAME is no longer running (status: $CONTAINER_STATUS)"

        # Get exit code if available
        EXIT_CODE=$(docker inspect "$CONTAINER_NAME" 2>/dev/null | jq -r '.[0].State.ExitCode' 2>/dev/null || echo "unknown")
        echo "Container exit code: $EXIT_CODE"

        # Update status file for local deployment
        if [ "$DEPLOYMENT_MODE" = "local" ] && [ -n "$STATUS_FILE" ]; then
            cat > "$STATUS_FILE" <<EOF
{
  "container_id": "$CONTAINER_ID",
  "container_name": "$CONTAINER_NAME",
  "container_image": "$CONTAINER_IMAGE",
  "started_at": "$STARTED_AT",
  "last_updated": "$(date -Iseconds)",
  "status": "exited",
  "exit_code": "$EXIT_CODE",
  "enclave_tools_exit_code": $ENCLAVE_TOOLS_EXIT_CODE
}
EOF
        fi
        break
    fi

    # Log current status periodically
    if [ $((MONITOR_COUNT % 10)) -eq 0 ]; then
        echo "$(date): Container $CONTAINER_NAME status: $CONTAINER_STATUS"

        # Show container resource usage
        docker stats --no-stream "$CONTAINER_NAME" 2>/dev/null || echo "Could not get container stats"

        # Update status file for local deployment
        if [ "$DEPLOYMENT_MODE" = "local" ] && [ -n "$STATUS_FILE" ]; then
            cat > "$STATUS_FILE" <<EOF
{
  "container_id": "$CONTAINER_ID",
  "container_name": "$CONTAINER_NAME",
  "container_image": "$CONTAINER_IMAGE",
  "started_at": "$STARTED_AT",
  "last_updated": "$(date -Iseconds)",
  "status": "$CONTAINER_STATUS",
  "monitor_count": $MONITOR_COUNT,
  "enclave_tools_exit_code": $ENCLAVE_TOOLS_EXIT_CODE
}
EOF
        fi
    fi

    MONITOR_COUNT=$((MONITOR_COUNT + 1))
    # sleep in the background and `wait` for it.  `wait` is a shell
    # builtin that is interrupted immediately by trapped signals, unlike
    # a foreground `sleep` which defers signal delivery until it returns.
    sleep "$MONITOR_INTERVAL" &
    wait $! 2>/dev/null || true
done

echo "Enclave monitoring ended"

# Clean up log capture if still running
if kill -0 $LOG_PID 2>/dev/null; then
    echo "Stopping log capture..."
    kill $LOG_PID 2>/dev/null || true
fi

echo "Script exiting..."
