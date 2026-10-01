# L1 Geth Dockerfile, modified from ops/docker/deployment-utils/Dockerfile
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    gcc \
    g++ \
    musl-dev \
    linux-headers

# Build eth-beacon-genesis
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    GOTOOLCHAIN=auto go install -ldflags '-linkmode external -extldflags "-static"' \
    github.com/ethpandaops/eth-beacon-genesis/cmd/eth-beacon-genesis@703e97a

# Build eth2-val-tools
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    go install -ldflags '-linkmode external -extldflags "-static"' \
    github.com/protolambda/eth2-val-tools@1393104

# Main runtime image
FROM debian:12.7-slim

ENV DEBIAN_FRONTEND=noninteractive

# Install runtime dependencies (coreutils provides sha256sum)
RUN apt-get update && apt-get install -y \
    coreutils \
    curl \
    jq \
    ca-certificates \
    openssl \
    && rm -rf /var/lib/apt/lists/*

# Install dasel for JSON manipulation
RUN curl -sSL "https://github.com/TomWright/dasel/releases/download/v2.8.1/dasel_linux_$(dpkg --print-architecture)" -o /usr/local/bin/dasel && \
    chmod +x /usr/local/bin/dasel

# Install yq for YAML parsing
RUN curl -sSL "https://github.com/mikefarah/yq/releases/latest/download/yq_linux_$(dpkg --print-architecture)" -o /usr/local/bin/yq && \
    chmod +x /usr/local/bin/yq

# Install Geth for the given architecture.
RUN ARCH=$(dpkg --print-architecture) && \
    echo "Detected architecture: $ARCH" && \
    case "$ARCH" in \
    "amd64") \
    GETH_URL="https://gethstore.blob.core.windows.net/builds/geth-linux-amd64-1.15.11-36b2371c.tar.gz" && \
    GETH_SHA="a14a4285daedf75ea04a7a298e6caa48d566a2786c93fc5e86ec2c5998c92455" && \
    GETH_DIR="geth-linux-amd64-1.15.11-36b2371c" && \
    VERIFY_SHA="true" \
    ;; \
    "arm64") \
    GETH_URL="https://gethstore.blob.core.windows.net/builds/geth-linux-arm64-1.15.11-36b2371c.tar.gz" && \
    GETH_SHA="148ec84db2268fa846ae68f6445f0c98d33e95069e40fe8c74b43ea5eb53df7b" && \
    GETH_DIR="geth-linux-arm64-1.15.11-36b2371c" && \
    VERIFY_SHA="true" \
    ;; \
    *) \
    echo "Unsupported architecture: $ARCH" && exit 1 \
    ;; \
    esac && \
    echo "Downloading: $GETH_URL" && \
    curl -L "$GETH_URL" -o geth.tar.gz && \
    echo "$GETH_SHA geth.tar.gz" | sha256sum -c - && \
    tar -xvf geth.tar.gz && \
    mv "$GETH_DIR/geth" /usr/local/bin/geth && \
    rm -rf geth.tar.gz "$GETH_DIR" && \
    chmod +x /usr/local/bin/geth

# Install lighthouse consensus tools
RUN ARCH=$(dpkg --print-architecture) && \
    case "$ARCH" in \
    "amd64") LH_ARCH="x86_64-unknown-linux-gnu" ;; \
    "arm64") LH_ARCH="aarch64-unknown-linux-gnu" ;; \
    *) echo "Unsupported architecture: $ARCH" && exit 1 ;; \
    esac && \
    curl -sSL "https://github.com/sigp/lighthouse/releases/download/v5.3.0/lighthouse-v5.3.0-${LH_ARCH}.tar.gz" -o lighthouse.tar.gz && \
    tar -xzf lighthouse.tar.gz && \
    mv lighthouse /usr/local/bin/ && \
    rm lighthouse.tar.gz && \
    chmod +x /usr/local/bin/lighthouse

# Copy binaries from builder stage
COPY --from=builder /go/bin/eth-beacon-genesis /usr/local/bin/eth-beacon-genesis
COPY --from=builder /go/bin/eth2-val-tools /usr/local/bin/eth2-val-tools

# Create data and templates directories
RUN mkdir -p /data /templates

# Include the initialization scripts for the L1 services.
COPY espresso/docker/l1-geth/beacon-config.yaml /templates/
COPY espresso/docker/l1-geth/devnet-genesis-template.json /templates/
COPY espresso/docker/l1-geth/mnemonics.yaml /templates/
COPY espresso/docker/l1-geth/l1-geth-init.sh /l1-geth-init.sh
# Deployment artifacts are mounted via docker-compose volumes at runtime
# (./deployment:/deployment) rather than baked into the image.

# Run the initialization script.
RUN chmod +x /l1-geth-init.sh

# Expose the RPC ports
EXPOSE 8545 8551

# Set working directory
WORKDIR /data

# Use the initialization script as the entrypoint.
ENTRYPOINT "/l1-geth-init.sh"
