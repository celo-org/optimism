FROM --platform=linux/amd64 us-west1-docker.pkg.dev/blockchaintestsglobaltestnet/dev-images/op-geth@sha256:6fe04ace89e6ebae1c527c2754d1eb40bcb073a85b2f23384816c53d94bbaff2

RUN apk add --no-cache jq

COPY l2-op-geth-entrypoint.sh /entrypoint.sh

VOLUME ["/db"]

ENTRYPOINT ["/bin/sh", "/entrypoint.sh"]
