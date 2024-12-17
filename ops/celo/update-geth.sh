#!/bin/bash

branch="$1"
if [ -z "$branch" ]; then
  echo "No argument given. Please supply the ref in 'celo-org/op-geth' to be used "
  exit 1
fi

commit=$(git ls-remote https://github.com/celo-org/op-geth/ "$branch" | awk '{print $1}')
if [ -z "$commit" ]; then exit 1; fi

go_version=$(go list -m "github.com/celo-org/op-geth@$commit")
if [ -z "$go_version" ]; then exit 1; fi

sha256digest=$(gcloud --format=json artifacts files list \
  --project=blockchaintestsglobaltestnet \
  --repository=dev-images \
  --location=us-west1 \
  --package=op-geth \
  --limit=1 \
  --tag="$commit" | jq ".[0].name" | grep -oP "sha256:\K[0-9a-f]{64}")
if [ -z "$sha256digest" ]; then exit 1; fi

sed -i "s|\(.*op-geth@sha256:\)\(.*\)|\1$sha256digest|" ./ops-bedrock/l2-op-geth.Dockerfile
sed -i "s|\(replace github.com/ethereum/go-ethereum => \)github.com/celo-org/op-geth v.*|\1$go_version|" go.mod

go_mod_error=$(go mod tidy >/dev/null)
if [ -n "$go_mod_error" ]; then
  echo "$go_mod_error"
  exit 1
fi

echo "$commit"
