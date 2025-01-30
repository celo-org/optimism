FROM golang:1.22.7-alpine3.20 AS builder
RUN apk add --no-cache curl tar gzip make gcc musl-dev linux-headers git jq bash

# install yq
RUN wget https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 -O /usr/local/bin/yq && \
    chmod +x /usr/local/bin/yq

COPY ./mise.toml .
RUN curl -L https://github.com/casey/just/releases/download/$(yq '.tools.just' mise.toml)/just-$(yq '.tools.just' mise.toml)-x86_64-unknown-linux-musl.tar.gz | \
    tar xz -C /usr/local/bin just

COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

WORKDIR /app

RUN go mod download

COPY ./op-service /app/op-service
COPY ./op-node /app/op-node
COPY ./op-alt-da /app/op-alt-da
COPY ./op-chain-ops /app/op-chain-ops
COPY ./justfiles /app/justfiles

WORKDIR /app/op-chain-ops
RUN just celo-migrate

FROM alpine:3.20
RUN apk --no-cache add ca-certificates bash rsync

# RUN addgroup -S app && adduser -S app -G app
# USER app
WORKDIR /app

COPY --from=builder /app/op-chain-ops/bin/celo-migrate /app
ENV PATH="/app:${PATH}"

ENTRYPOINT ["/app/celo-migrate"]
