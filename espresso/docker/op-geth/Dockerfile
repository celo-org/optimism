# OP Geth Dockerfile

# Build arguments.
ARG TARGETOS
ARG TARGETARCH
ARG GIT_COMMIT
ARG GIT_DATE

# CGO builder for components that need Espresso crypto linking.
FROM golang:1.26.4-alpine AS op-cgo-builder
# Install dependencies
RUN apk add musl-dev gcc g++ curl tar gzip make linux-headers git jq bash

# Go sources
COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum
# Warm-up the cache
WORKDIR /app
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build go mod download
COPY . /app
COPY ./op-validator/ /app/op-validator/

# Build op-deployer
FROM op-cgo-builder AS op-deployer-builder
ARG OP_DEPLOYER_VERSION=v0.0.0
ENV GOOS=$TARGETOS GOARCH=$TARGETARCH GITCOMMIT=$GIT_COMMIT GITDATE=$GIT_DATE VERSION="$OP_DEPLOYER_VERSION"
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
  go build -ldflags '-linkmode external -extldflags "-static"' -o /op-deployer \
  ./op-deployer/cmd/op-deployer

# Build geth from Celo fork (go.mod replaces github.com/ethereum/go-ethereum with github.com/celo-org/op-geth)
# Required because the devnet uses Jovian eip-1559 params (17-byte extraData) which the OP Labs geth does not support.
# GONOSUMDB=* and GOFLAGS=-mod=mod allow resolving cmd/geth dependencies not in our go.sum.
FROM op-cgo-builder AS geth-builder
ENV GONOSUMDB="*" GOFLAGS="-mod=mod"
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
  go build -ldflags '-linkmode external -extldflags "-static"' -o /geth github.com/ethereum/go-ethereum/cmd/geth

FROM us-docker.pkg.dev/oplabs-tools-artifacts/images/op-geth:v1.101503.2-rc.3

# Replace OP Labs geth with Celo fork geth (supports Jovian eip-1559 extraData format).
COPY --from=geth-builder /geth /usr/local/bin/geth

# For healtcheck and JSON operations.
RUN apk add curl jq openssl bash

# Install dasel for JSON manipulation.
RUN case $(uname -m) in \
  "arm64"|"aarch64") ARCH="arm64" ;; \
  *) ARCH="amd64" ;; \
  esac && \
  curl -sSL "https://github.com/TomWright/dasel/releases/download/v2.8.1/dasel_linux_$ARCH" -o /usr/local/bin/dasel && \
  chmod +x /usr/local/bin/dasel

# Copy binary from builder stage.
COPY --from=op-deployer-builder /op-deployer /usr/local/bin/

# Deployment artifacts are mounted via docker-compose volumes at runtime
# (./deployment:/deployment, ./deployment/deployer:/deployer, etc.)
# rather than baked into the image.
COPY espresso/docker/op-geth/op-geth-init.sh /op-geth-init.sh

# Run the initialization script.
RUN chmod +x /op-geth-init.sh

# Create data directory.
RUN mkdir -p /data

# Expose the RPC ports.
EXPOSE 8546 8552

# Set working directory.
WORKDIR /data

# Use the initialization script as the entrypoint.
ENTRYPOINT ["/op-geth-init.sh"]
