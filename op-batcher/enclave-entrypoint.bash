#!/usr/bin/env bash

# Entrypoint for op-batcher running in enclaver image.
# Uses HTTPS_PROXY for external URLs (preserving SNI/Host headers)
# Only rewrites internal localhost URLs to map to enclave's "host"

set -e

echo "=== Enclave Environment Debug Info ==="
echo "PATH: $PATH"
echo "Working directory: $(pwd)"
echo "Proxy: ${http_proxy:-not set}"
echo "======================================"

# Re-populate the arguments passed through the environment
if [ -n "$ENCLAVE_BATCHER_ARGS" ]; then
  eval set -- "$ENCLAVE_BATCHER_ARGS"
fi

# Store the original arguments from ENCLAVE_BATCHER_ARGS
original_args=("$@")

# ---------------------------------------------------------------------------
# Start nc listener in background IMMEDIATELY — before Odyn setup.
#
# ---------------------------------------------------------------------------
# These port numbers must match ArgDeliveryPort / ReadinessPort in enclave-tools/enclaver.go.
NC_PORT=8337
READY_PORT=8338
received_args=()
NC_TMPFILE=$(mktemp)
echo "nc listener starting on port $NC_PORT (background, 60 second timeout)"
nc -l -p "$NC_PORT" -w 60 > "$NC_TMPFILE" &
NC_BG_PID=$!
# Ensure nc process and tempfile are cleaned up on any exit path.
# kill/rm are no-ops if the process/file is already gone.
trap 'kill "$NC_BG_PID" 2>/dev/null; rm -f "$NC_TMPFILE"' EXIT
if ! kill -0 "$NC_BG_PID" 2>/dev/null; then
    echo "[ERROR] nc listener failed to start on port $NC_PORT" >&2
    exit 1
fi
echo "nc listener running (PID $NC_BG_PID)"

# Verify Odyn proxy is available
if [ -z "$http_proxy" ]; then
    echo "[ERROR] http_proxy not set" >&2
    exit 1
fi

if ! ODYN_PROXY_PORT=$(trurl --url "$http_proxy" --get "{port}"); then
    echo "[ERROR] Failed to parse http_proxy" >&2
    exit 1
fi

echo "[DEBUG] Testing Odyn proxy on port $ODYN_PROXY_PORT..." >&2
if nc -z 127.0.0.1 $ODYN_PROXY_PORT 2>/dev/null; then
  echo "✓ Odyn proxy functional on port $ODYN_PROXY_PORT"
else
  echo "[ERROR] Odyn proxy unreachable on port $ODYN_PROXY_PORT" >&2
  exit 1
fi

# CRITICAL: Preserve proxy environment variables for Go's HTTP client
# This allows external HTTPS URLs to work with correct SNI and Host headers
export HTTPS_PROXY="$http_proxy"
export HTTP_PROXY="$http_proxy"
export https_proxy="$http_proxy"
export NO_PROXY="localhost,127.0.0.1,::1,host"
export no_proxy="$NO_PROXY"

echo "[DEBUG] Proxy environment configured:"
echo "  HTTPS_PROXY=$HTTPS_PROXY"
echo "  NO_PROXY=$NO_PROXY"
echo "[DEBUG] External URLs will use proxy with correct SNI/Host headers"
echo ""

# Readiness handshake — send "READY" on port $READY_PORT (background).
#
# Sending READY here (after the Odyn check) proves two things to the caller:
# That nc:$NC_PORT is already listening and Odyn egress proxy is verified and functional
echo "Signaling readiness on port $READY_PORT (background)..."
echo "READY" | nc -l -p "$READY_PORT" -w 30 &

# Helper function to check if URL needs socat proxying
# URLs pointing to localhost, 127.0.0.1, or "host" need socat because:
# - Go's HTTP client cannot resolve "host" hostname via DNS
# - These are internal enclave connections that need special handling
needs_socat_proxy() {
    local url="$1"
    local host
    host="$(trurl --url "$url" --get "{host}" 2>/dev/null)" || return 1

    if [[ "$host" == "localhost" ]] || [[ "$host" == "127.0.0.1" ]] || [[ "$host" == "::1" ]] || [[ "$host" == "host" ]]; then
        return 0  # needs socat
    fi
    return 1  # is external, use HTTPS_PROXY
}

# Helper function to wait for socat to open a port
wait_for_port() {
    local port="$1"

    for ((i=0; i<100; i++)); do
        if nc -z 127.0.0.1 "$port" 2>/dev/null; then
            return 0
        fi
        sleep 0.3
    done

    echo "[ERROR] socat did not open port $port in time" >&2
    return 1
}

# Helper function to launch socat proxy for internal URLs
launch_socat() {
    local original_url="$1"
    local socat_port="$2"

    local host port scheme path
    if ! read -r host port scheme path < <(trurl --url "$original_url" --default-port --get "{host} {port} {scheme} {path}"); then
        echo "[ERROR] Failed to parse URL: $original_url" >&2
        return 1
    fi

    # Map localhost to "host" for enclave's parent
    if [[ "$host" == "localhost" ]] || [[ "$host" == "127.0.0.1" ]] || [[ "$host" == "::1" ]]; then
        echo "[DEBUG] Rewriting '$host' to 'host'" >&2
        host="host"
    fi

    if [[ "$scheme" != "http" ]] && [[ "$scheme" != "https" ]]; then
        echo "[ERROR] Invalid scheme: '$scheme'. Only http and https are supported." >&2
        return 1
    fi

    # Start socat to proxy through Odyn to "host"
    echo "[DEBUG] Starting socat: 127.0.0.1:${socat_port} -> PROXY:${host}:${port} via Odyn:${ODYN_PROXY_PORT}" >&2
    socat -t 10 -d TCP4-LISTEN:"${socat_port}",reuseaddr,fork PROXY:127.0.0.1:"$host":"$port",proxyport="${ODYN_PROXY_PORT}" > /dev/null 2>&1 &
    socat_pid=$!
    disown "$socat_pid"

    wait_for_port "${socat_port}" || {
        kill "$socat_pid" 2>/dev/null
        wait "$socat_pid" 2>/dev/null
        return 1
    }

    # Return socat-proxied URL (strip trailing slash to avoid double-slash in paths)
    local new_url
    new_url="$(trurl --url "$original_url" --set host="127.0.0.1" --set port="$socat_port")"
    # Remove trailing slash if present
    new_url="${new_url%/}"
    echo "$new_url"

    return 0
}

# Wait for background nc to finish — by this point it has either already
# received all args (connection came in during Odyn setup) or we block up to
# the 60 second nc timeout.
echo "Waiting for batcher arguments (nc PID $NC_BG_PID)..."
wait "$NC_BG_PID" 2>/dev/null || true

if [ -s "$NC_TMPFILE" ]; then
    while IFS= read -r -d '' arg; do
        if [[ -z "$arg" ]]; then
            break
        fi
        received_args+=("$arg")
    done < "$NC_TMPFILE"
fi
rm -f "$NC_TMPFILE"

if [ ${#received_args[@]} -eq 0 ]; then
    echo "Warning: No arguments received via nc listener within 60 seconds, using original arguments"
    set -- "${original_args[@]}"
else
    echo "Received ${#received_args[@]} arguments via nc, merging with original arguments"
    set -- "${original_args[@]}" "${received_args[@]}"
fi

# URL argument regex pattern
URL_ARG_RE='^(--altda\.da-server|--espresso\.espresso-attestation-service|--espresso\.urls|--espresso\.l1-url|--espresso\.rollup-l1-url|--l1-eth-rpc|--l2-eth-rpc|--rollup-rpc|--signer\.endpoint)(=|$)'
# Process all arguments
filtered_args=()
url_args=()
SOCAT_PORT=10001

echo "Processing arguments..."
while [ $# -gt 0 ]; do
    # Check if the argument matches the URL pattern
    if [[ $1 =~ $URL_ARG_RE ]]; then
        flag=${BASH_REMATCH[1]}

        # Extract value from "--flag=value" or "--flag value"
        if [[ "$1" == *=* ]]; then
            value="${1#*=}"
        else
            shift || { echo "$flag missing value"; exit 1; }
            value="$1"
        fi

        # Handle comma-separated values for any flag
        if [[ "$value" == *","* ]]; then
            IFS=',' read -r -a parts <<< "$value"
            rewritten_parts=()
            for part in "${parts[@]}"; do
                if needs_socat_proxy "$part"; then
                    if ! new_url=$(launch_socat "$part" "$SOCAT_PORT"); then
                        echo "[ERROR] Failed to launch socat for $flag=$part" >&2
                        exit 1
                    fi
                    echo "[DEBUG] Proxying internal URL via socat: $part -> $new_url" >&2
                    rewritten_parts+=("$new_url")
                    SOCAT_PORT=$((SOCAT_PORT + 1))
                else
                    echo "[DEBUG] Keeping external URL unchanged (will use HTTPS_PROXY): $part" >&2
                    rewritten_parts+=("$part")
                fi
            done
            # Join with commas
            joined=$(IFS=,; echo "${rewritten_parts[*]}")
            url_args+=("${flag}=${joined}")
        else
            if needs_socat_proxy "$value"; then
                if ! new_url=$(launch_socat "$value" "$SOCAT_PORT"); then
                    echo "[ERROR] Failed to launch socat for $flag=$value" >&2
                    exit 1
                fi
                echo "[DEBUG] Proxying internal URL via socat: $value -> $new_url" >&2
                url_args+=("$flag" "$new_url")
                SOCAT_PORT=$((SOCAT_PORT + 1))
            else
                echo "[DEBUG] Keeping external URL unchanged (will use HTTPS_PROXY): $value" >&2
                url_args+=("$flag" "$value")
            fi
        fi
    else
        filtered_args+=("$1")
    fi
    shift
done

# Combine all arguments
all_args=("${filtered_args[@]}" "${url_args[@]}")

echo ""
echo "=== Final op-batcher arguments ==="
echo "Total arguments: ${#all_args[@]}"
next_is_secret=false
for i in "${!all_args[@]}"; do
    arg="${all_args[$i]}"
    if $next_is_secret; then
        echo "  [$i]: [REDACTED]" >&2
        next_is_secret=false
    elif [[ "$arg" =~ ^--(private-key|mnemonic)= ]]; then
        flag="${arg%%=*}"
        echo "  [$i]: ${flag}=[REDACTED]" >&2
    elif [[ "$arg" == "--private-key" || "$arg" == "--mnemonic" ]]; then
        echo "  [$i]: $arg" >&2
        next_is_secret=true
    else
        echo "  [$i]: $arg" >&2
    fi
done
echo "===================================" >&2
echo ""

echo "[DEBUG] Launching op-batcher..." >&2
exec op-batcher "${all_args[@]}"
