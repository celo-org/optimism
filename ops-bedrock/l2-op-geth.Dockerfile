a87d32edb91c6419a34807b96bd853304ea4f62671db9dd33aceb2be8d6cffe

RUN apk add --no-cache jq

COPY l2-op-geth-entrypoint.sh /entrypoint.sh

VOLUME ["/db"]

ENTRYPOINT ["/bin/sh", "/entrypoint.sh"]
