#!/bin/bash

branch="${1:-celo10}"
commit=$(git ls-remote https://github.com/celo-org/op-geth/ $branch | awk '{print $1}')
go_version=$(go list -m "github.com/celo-org/op-geth@$commit")
sha256digest=$(gcloud --format=json artifacts files list \
  --project=blockchaintestsglobaltestnet \
  --repository=dev-images \
  --location=us-west1 \
  --package=op-geth \
  --limit=1 \
  --tag=$commit | jq ".[0].name" | grep -oP "sha256:\K[0-9a-f]{64}")

sed -i "s|\(.*op-geth@sha256:\)\(.*\)|\1$sha256digest|" ./ops-bedrock/l2-op-geth.Dockerfile
sed -i "s|\(replace github.com/ethereum/go-ethereum => \)github.com/celo-org/op-geth v.*|\1$go_version|" go.mod
go mod tidy

files="go.mod go.sum ops-bedrock/l2-op-geth.Dockerfile"
if git diff --quiet HEAD -- $files; then
  echo "No updates."
else
  echo "Updates, comitting"
  git add $files
  printf "update: \`celo/op-geth\`\n

 Update the go package dependency and the devnet
 docker container reference of the \`l2\` service
 to the latest commit ('%s') in the '%s' ref." "$commit" "$branch" | git commit -F -
fi
