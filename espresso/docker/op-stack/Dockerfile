# OP Stack Dockerfile, simplified from ops/docker/op-stack-go/Dockerfile

# Build arguments
ARG TARGET_BASE_IMAGE=alpine:3.22
ARG TARGETOS
ARG TARGETARCH

# Base builder image. Keep this in sync with the root go.mod toolchain.
FROM golang:1.26.4-alpine AS builder
# Add ARG to fix arm64 build.
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

RUN apk add --no-cache \
  curl netcat-openbsd tar gzip make gcc g++ musl-dev \
  linux-headers git bash jq

# Install yq, dasel, and foundry
RUN case "$TARGETARCH" in \
  "amd64") ARCH="amd64" ;; \
  "arm64") ARCH="arm64" ;; \
  *) ARCH="amd64" ;; \
  esac && \
  wget https://github.com/mikefarah/yq/releases/latest/download/yq_linux_$ARCH -O /usr/local/bin/yq && \
  chmod +x /usr/local/bin/yq && \
  wget https://github.com/TomWright/dasel/releases/download/v2.8.1/dasel_linux_$ARCH -O /usr/local/bin/dasel && \
  chmod +x /usr/local/bin/dasel && \
  wget https://github.com/foundry-rs/foundry/releases/download/v1.4.4/foundry_v1.4.4_linux_$ARCH.tar.gz -O foundry.tgz && \
  tar -xzf foundry.tgz -C /usr/local/bin && \
  chmod +x /usr/local/bin/chisel && \
  chmod +x /usr/local/bin/anvil && \
  chmod +x /usr/local/bin/cast && \
  chmod +x /usr/local/bin/forge

# Install `just` without hitting GitHub (avoids rate limits in CI).
RUN apk add --no-cache just && just --version

# Ensure just and other tools are on PATH for all FROM builder stages
ENV PATH="/usr/local/bin:$PATH"


# Copy and download Go dependencies
COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum
WORKDIR /app
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build go mod download

# Copy source code
COPY . /app
COPY ./op-validator/ /app/op-validator/

# Build arguments for git metadata
ARG GIT_COMMIT
ARG GIT_DATE

# Build op-node
FROM builder AS op-node-builder
ARG TARGETOS
ARG TARGETARCH
ARG GIT_COMMIT
ARG GIT_DATE
ARG OP_NODE_VERSION=v0.0.0
ENV GOOS=$TARGETOS GOARCH=$TARGETARCH GITCOMMIT=$GIT_COMMIT GITDATE=$GIT_DATE VERSION="$OP_NODE_VERSION"
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
  cd /app/op-node && mkdir -p bin && \
  CGO_ENABLED=0 go build -ldflags '-extldflags "-static" -X main.GitCommit=$GITCOMMIT -X main.GitDate=$GITDATE -X github.com/ethereum-optimism/optimism/op-node/version.Version=$VERSION -X github.com/ethereum-optimism/optimism/op-node/version.Meta=' -o bin/op-node ./cmd/main.go

# Build op-batcher
FROM builder AS op-batcher-builder
ARG TARGETOS
ARG TARGETARCH
ARG GIT_COMMIT
ARG GIT_DATE
ARG OP_BATCHER_VERSION=v0.0.0
ENV GOOS=$TARGETOS GOARCH=$TARGETARCH GITCOMMIT=$GIT_COMMIT GITDATE=$GIT_DATE VERSION="$OP_BATCHER_VERSION"
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
  cd /app/op-batcher && mkdir -p bin && go build -v -ldflags "-X main.GitCommit=$GITCOMMIT -X main.GitDate=$GITDATE -X main.Version=$VERSION" -o ./bin/op-batcher ./cmd

# Build enclave-tools
FROM builder AS enclave-tools-builder
ARG TARGETOS
ARG TARGETARCH
ARG GIT_COMMIT
ARG GIT_DATE
ARG ENCLAVE_TOOLS_VERSION=v0.0.0
ENV GOOS=$TARGETOS GOARCH=$TARGETARCH GITCOMMIT=$GIT_COMMIT GITDATE=$GIT_DATE VERSION="$ENCLAVE_TOOLS_VERSION"
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
  cd /app/op-batcher && mkdir -p bin && go build -v -o ./bin/enclave-tools ./enclave-tools/cmd

# Build op-proposer
FROM builder AS op-proposer-builder
ARG TARGETOS
ARG TARGETARCH
ARG GIT_COMMIT
ARG GIT_DATE
ARG OP_PROPOSER_VERSION=v0.0.0
ENV GOOS=$TARGETOS GOARCH=$TARGETARCH GITCOMMIT=$GIT_COMMIT GITDATE=$GIT_DATE VERSION="$OP_PROPOSER_VERSION"
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build cd /app/op-proposer && make op-proposer
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build cd /app/op-challenger && make op-challenger

# Build op-deployer
FROM builder AS op-deployer-builder
ARG TARGETOS
ARG TARGETARCH
ARG GIT_COMMIT
ARG GIT_DATE
ARG OP_DEPLOYER_VERSION=v0.0.0
ENV GOOS=$TARGETOS GOARCH=$TARGETARCH GITCOMMIT=$GIT_COMMIT GITDATE=$GIT_DATE VERSION="$OP_DEPLOYER_VERSION"
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
  cd /app/op-deployer && just build-go


# Download Aztec SRS once; COPY --from into each target that needs it.
FROM $TARGET_BASE_IMAGE AS aztec-srs
ADD "https://github.com/EspressoSystems/ark-srs/releases/download/v0.2.0/kzg10-aztec20-srs-1048584.bin" /aztec/kzg10-aztec20-srs-1048584.bin

# Final runtime images
FROM $TARGET_BASE_IMAGE AS op-node-target
RUN apk add libgcc curl
ENV AZTEC_SRS_PATH=/aztec/kzg10-aztec20-srs-1048584.bin
COPY --from=aztec-srs /aztec/kzg10-aztec20-srs-1048584.bin /aztec/kzg10-aztec20-srs-1048584.bin
COPY --from=op-node-builder /app/op-node/bin/op-node /usr/local/bin/

# Create config directory
RUN mkdir -p /config

# Deployment artifacts are mounted via docker-compose volumes at runtime
# (./deployment/l2-config:/config, ./deployment/deployer:/deployer, etc.)
# rather than baked into the image, so that runtime-generated files in
# espresso/deployment/ don't bust the Docker build cache.

CMD ["op-node"]

FROM $TARGET_BASE_IMAGE AS op-batcher-target
RUN apk add libgcc
ENV AZTEC_SRS_PATH=/aztec/kzg10-aztec20-srs-1048584.bin
COPY --from=aztec-srs /aztec/kzg10-aztec20-srs-1048584.bin /aztec/kzg10-aztec20-srs-1048584.bin
COPY --from=op-batcher-builder /app/op-batcher/bin/op-batcher /usr/local/bin/
CMD ["op-batcher"]

# Inner image that Enclaver wraps into the EIF. Contains only the batcher
# binary and the enclave entrypoint (socat proxying, arg routing via nc).
FROM $TARGET_BASE_IMAGE AS op-batcher-enclave-app
RUN apk add libgcc bash socat trurl
ENV AZTEC_SRS_PATH=/aztec/kzg10-aztec20-srs-1048584.bin
COPY --from=aztec-srs /aztec/kzg10-aztec20-srs-1048584.bin /aztec/kzg10-aztec20-srs-1048584.bin
COPY --from=op-batcher-builder /app/op-batcher/bin/op-batcher /usr/local/bin/
COPY op-batcher/enclave-entrypoint.bash /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
ENTRYPOINT ["/app/entrypoint.sh"]

# Runtime image for TEE batcher.  Uses `enclave-tools build-eif` with the
# pre-built op-batcher-enclave-app image (eliminating a full rebuild from
# source at startup) then runs the resulting EIF via `enclave-tools run`.
FROM $TARGET_BASE_IMAGE AS op-batcher-enclave-target
RUN apk add libgcc docker bash jq curl wget
# Install enclaver for EIF creation
RUN curl -L https://github.com/enclaver-io/enclaver/releases/download/v0.6.1/enclaver-linux-x86_64-v0.6.1.tar.gz | tar xz --strip-components=1 -C /usr/local/bin enclaver-linux-x86_64-v0.6.1/enclaver
WORKDIR /app
# Deployment state (state.json with BatchAuthenticatorAddress) is mounted via
# docker-compose volumes at /app/deployment rather than baked into the image.
# run-enclave.sh — build EIF from pre-built app image + register PCR0 + run
COPY espresso/docker/op-batcher-tee/run-enclave.sh /app/run-enclave.sh
RUN chmod +x /app/run-enclave.sh
COPY --from=enclave-tools-builder /app/op-batcher/bin/enclave-tools /usr/local/bin/
CMD ["enclave-tools"]

FROM $TARGET_BASE_IMAGE AS op-proposer-target
RUN apk add jq
COPY espresso/docker/op-stack/entrypoint.sh /bin/entrypoint.sh
RUN chmod +x /bin/entrypoint.sh
COPY --from=op-proposer-builder /app/op-proposer/bin/op-proposer /usr/local/bin/
# /deployer is mounted via docker-compose volumes (./deployment/deployer:/deployer)
ENV ENV_PREFIX=OP_PROPOSER
ENTRYPOINT [ "/bin/entrypoint.sh" ]
CMD ["op-proposer"]

FROM $TARGET_BASE_IMAGE AS op-challenger-target
RUN apk add jq
COPY espresso/docker/op-stack/entrypoint.sh /bin/entrypoint.sh
RUN chmod +x /bin/entrypoint.sh
COPY --from=op-proposer-builder /app/op-challenger/bin/op-challenger /usr/local/bin
# Note: cannon, op-program, and prestate-proof.json are not built in this Dockerfile
# They would require separate builder stages with CGO support
# COPY --from=op-proposer-builder /app/cannon/bin/cannon /usr/local/bin
# COPY --from=op-proposer-builder /app/op-program/bin/op-program /usr/local/bin
# COPY --from=op-proposer-builder /app/op-program/bin/prestate-proof.json /app/prestate-proof.json

# ENV OP_CHALLENGER_CANNON_BIN /usr/local/bin/cannon
# ENV OP_CHALLENGER_CANNON_SERVER /usr/local/bin/op-program
# ENV OP_CHALLENGER_CANNON_PRESTATE /app/prestate-proof.json
ENV ENV_PREFIX=OP_CHALLENGER

ENTRYPOINT [ "/bin/entrypoint.sh" ]
CMD ["op-challenger"]

FROM $TARGET_BASE_IMAGE AS op-deployer-target
RUN apk add jq curl bash openssl
COPY --from=builder /usr/local/bin/dasel /usr/local/bin/
COPY --from=builder /usr/local/bin/yq /usr/local/bin/
COPY --from=builder /usr/local/bin/cast /usr/local/bin/
COPY --from=op-deployer-builder /app/op-deployer/bin/op-deployer /usr/local/bin
COPY /packages/contracts-bedrock/forge-artifacts /contracts
ENV DEPLOYER_ARTIFACTS_LOCATOR=/contracts
CMD ["op-deployer"]
