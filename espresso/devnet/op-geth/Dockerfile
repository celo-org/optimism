# op-geth for the Espresso compose devnet (espresso/docker-compose.yml).
#
# Builds geth from the module that go.mod replaces github.com/ethereum/go-ethereum
# with (celo-org/op-geth), so the devnet runs the same execution client the Go
# services are built against. Build context: the repository root.

FROM golang:1.26.4-alpine3.22 AS builder
RUN apk add --no-cache gcc musl-dev linux-headers
WORKDIR /app
COPY go.mod go.sum ./
# cmd/geth pulls in packages that the monorepo does not import, so their
# checksums are not in go.sum.
ENV GOFLAGS=-mod=mod GONOSUMDB=* GOTOOLCHAIN=local
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
  go build -ldflags '-linkmode external -extldflags "-static"' -o /geth github.com/ethereum/go-ethereum/cmd/geth

FROM alpine:3.22
RUN apk add --no-cache bash curl jq
COPY --from=builder /geth /usr/local/bin/geth
COPY --chmod=755 espresso/devnet/op-geth/op-geth-init.sh /op-geth-init.sh
RUN mkdir -p /data
WORKDIR /data
EXPOSE 8546 8552
ENTRYPOINT ["/op-geth-init.sh"]
