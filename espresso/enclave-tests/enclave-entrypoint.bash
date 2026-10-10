#!/usr/bin/env bash
# Entrypoint for op-batcher inside the enclave. Receives the batcher args over vsock and
# proxies localhost URLs to the enclave's parent through the Odyn egress proxy.
set -euo pipefail

# Must match enclaveArgsPort / enclaveReadyPort in enclave_helpers_test.go.
NC_PORT=8337
READY_PORT=8338

ARGS_FILE=$(mktemp)
nc -l -p "$NC_PORT" -w 60 > "$ARGS_FILE" &
NC_PID=$!

: "${http_proxy:?http_proxy not set}"
ODYN_PROXY_PORT=$(trurl --url "$http_proxy" --get "{port}")
nc -z 127.0.0.1 "$ODYN_PROXY_PORT"

export HTTPS_PROXY="$http_proxy" HTTP_PROXY="$http_proxy" https_proxy="$http_proxy"
export NO_PROXY="localhost,127.0.0.1,::1,host" no_proxy="localhost,127.0.0.1,::1,host"

echo "READY" | nc -l -p "$READY_PORT" -w 30 &

wait "$NC_PID" || true
args=()
while IFS= read -r -d '' arg && [ -n "$arg" ]; do
    args+=("$arg")
done < "$ARGS_FILE"
if [ ${#args[@]} -eq 0 ]; then
    echo "No batcher arguments received" >&2
    exit 1
fi

# Go's HTTP client can't reach the parent's localhost, so forward each local URL through
# socat to "host" via Odyn.
SOCAT_PORT=10001
for i in "${!args[@]}"; do
    if [[ "${args[$i]}" =~ ^(--[^=]+=)(https?://(localhost|127\.0\.0\.1)(:[0-9]+)?.*)$ ]]; then
        flag=${BASH_REMATCH[1]}
        url=${BASH_REMATCH[2]}
        port=$(trurl --url "$url" --default-port --get "{port}")
        socat -t 10 TCP4-LISTEN:"$SOCAT_PORT",reuseaddr,fork PROXY:127.0.0.1:host:"$port",proxyport="$ODYN_PROXY_PORT" > /dev/null 2>&1 &
        for _ in $(seq 100); do nc -z 127.0.0.1 "$SOCAT_PORT" && break; sleep 0.3; done
        new_url=$(trurl --url "$url" --set host=127.0.0.1 --set port="$SOCAT_PORT")
        args[i]="${flag}${new_url%/}"
        SOCAT_PORT=$((SOCAT_PORT + 1))
    fi
done

exec op-batcher "${args[@]}"
