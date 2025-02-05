#!/bin/bash
set -eo pipefail

branch="$1"
if [ -z "$branch" ]; then
  echo "No argument given. Please supply the ref in 'celo-org/op-geth' to be used " >&2
  exit 1
fi

commit=$(git ls-remote https://github.com/celo-org/op-geth/ "$branch" | awk '{print $1}')
if [ -z "$commit" ]; then
  echo "Could not find branch '$branch' in 'celo-org/op-geth'" >&2
  exit 1
fi

go_version=$(go list -m "github.com/celo-org/op-geth@$commit")
if [ -z "$go_version" ]; then
  echo "Failed to generate go version string fork '$commit' in 'celo-org/op-geth'" >&2
  exit 1
fi

sha256digest=$(gcloud --format=json artifacts files list \
  --project=blockchaintestsglobaltestnet \
  --repository=dev-images \
  --location=us-west1 \
  --package=op-geth \
  --limit=1 \
  --tag="$commit" | jq ".[0].name" | grep -oE 'sha256:([0-9a-f]{64})' | sed 's/^sha256://')
if [ -z "$sha256digest" ]; then
  echo "Failed to find sha256digest for op-geth docker image 'celo-org/op-geth'" >&2
  exit 1
fi

# We need to escape the '@' otherwise '@sha256' is interpreted as a global
# Symbol by perl.
docker_search_string="(.*op-geth\@sha256:)(.*)"
gomod_search_string="^(replace github.com/ethereum/go-ethereum .*=> )github.com/.*/op-geth v.*"

# Check that the searches are each matching a single line
if [ "$(perl -ne "m|${docker_search_string}| && print" ops-bedrock/l2-op-geth.Dockerfile | wc -l)" != "1" ]; then
  echo "Failed to find exactly one match for docker search string in ops-bedrock/l2-op-geth.Dockerfile" >&2
  exit 1
fi

if [ "$(perl -ne "m|${gomod_search_string}| && print" go.mod | wc -l)" != "1" ]; then
  echo "Failed to find exactly one match for go mod search string in go.mod" >&2
  exit 1
fi

# perl -pi -e "s|${docker_search_string}|\${1}${sha256digest}|" ops-bedrock/l2-op-geth.Dockerfile
perl -pi -e "s|${gomod_search_string}|\${1}${go_version}|" go.mod

go_mod_error=$(go mod tidy >/dev/null)
if [ -n "$go_mod_error" ]; then
  echo "$go_mod_error"
  exit 1
fi

echo "$commit"
