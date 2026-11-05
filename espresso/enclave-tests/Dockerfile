# Wraps an op-batcher image (e.g. op-batcher-target from ops/docker/op-stack-go/Dockerfile)
# with the entrypoint that receives its args inside a Nitro enclave.
ARG OP_BATCHER_IMAGE
FROM ${OP_BATCHER_IMAGE}
RUN apk add --no-cache bash socat trurl
COPY --chmod=755 enclave-entrypoint.bash /app/entrypoint.sh
ENTRYPOINT ["/app/entrypoint.sh"]
