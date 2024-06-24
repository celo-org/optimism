FROM --platform=linux/amd64 us-west1-docker.pkg.dev/blockchaintestsglobaltestnet/dev-images/op-geth@sha256:0f7500f796362a9b9102c28cec6d94a590e219777b8344eef3784f629b720bb9

RUN apk add --no-cache jq

COPY l2-op-geth-entrypoint.sh /entrypoint.sh

VOLUME ["/db"]

ENTRYPOINT ["/bin/sh", "/entrypoint.sh"]
