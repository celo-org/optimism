---
name: celo-rebase
description: Rebase Celo's fork onto a newer upstream OP Stack release. Covers choosing a base that matches the op-geth fork, history cleanup, the rebase pass with per-commit linting, conflict resolution (incl. CircleCI/workflows and modify/delete), op-geth dependency + contract snapshot/semver + rust Cargo.lock regeneration, whole-branch build/lint verification, and publishing the new celo-rebase-XX branch. Use when starting or continuing a Celo rebase, resolving rebase conflicts, or asked to "do the next rebase".
---

# Celo Rebase

Celo maintains its OP Stack changes as a clean, descriptive patch set on top of upstream.
Periodically we create a new `celo-rebase-XX` branch (incrementing `XX`) and replay our
changes onto the latest upstream release. This keeps our diff small and reviewable and makes
future rebases easier. A rebase is also the right time to clean up history.

The result is a true `git rebase`: `celo-rebase-XX` is a linear stack of our commits (no merge
commits) on top of a chosen upstream **release tag**. The companion branch
`celo-rebase-XX-upstream` records exactly which upstream commit we rebased onto.

## When to Use

- Starting the next rebase (`celo-rebase-(N+1)`) onto a newer upstream release
- Continuing/finishing an in-progress rebase (resolving conflicts, getting lint/tests green)

Trigger phrases: "do the next celo rebase", "rebase onto vX.Y.Z", "resolve these rebase
conflicts", "finish celo-rebase-XX".

## Conventions

- Remotes: `upstream` = `ethereum-optimism/optimism`, `origin` = `celo-org/optimism`.
- `celo-*` branch names are **protected**. Do all work on a `user/`-prefixed branch.
- Pick the upstream base as an **op-node release tag** and keep it in `$BASE` for the session,
  e.g. `export BASE=op-node/v1.17.0`.
- **op-geth and optimism are coupled — they must match (this drives the base choice).** Each
  op-node release pins a specific `ethereum-optimism/op-geth` version in `go.mod`
  (`replace github.com/ethereum/go-ethereum => github.com/ethereum-optimism/op-geth vX`). Celo's
  go.mod replaces go-ethereum with the celo op-geth fork, so the **celo op-geth fork must be
  rebased onto the same upstream op-geth** the chosen op-node tag pins. Pick the op-node tag whose
  op-geth pin matches the op-geth fork's base (or rebase op-geth to match) — otherwise nothing
  compiles. See §2 and §4.
- This is **two rebases**: op-geth (`celo-org/op-geth`) *first*, then optimism (this repo).

## Overview

1. Clean up our commit history (optional but recommended).
2. Choose the upstream commit (`$BASE`) — constrained by the op-geth fork (they must match).
3. Rebase in passes: conflicts + lint first, then tests, then final cleanups.
4. Handle repo-specific tasks (op-geth dep first, then contract snapshots/semver, rust Cargo.lock,
   CI workflows) and verify the whole branch builds + lints (§3).
5. Publish: set `celo-rebase-XX-upstream` to `$BASE` and push `celo-rebase-XX` (do the same for
   `celo-org/op-geth`), then announce.

## 1. Clean up history

Rebasing is easier with a small, clean, descriptive set of commits. Ideally each commit looks
like it was written perfectly on the first try, directly on top of upstream. Typical cleanups:

- Reorder commits so related changes form one group.
- Fold reverts into the commit that introduced the code.
- Edit history so renames use the correct name from the start (remove the rename commit).
- Squash commits with a lot of back-and-forth.
- Drop commits that were only needed temporarily.

Aim for: linters pass and tests succeed **after every commit** (see "Avoiding breakage").

## 2. Choose the upstream commit

For finding the tag to rebase onto, use `git tag -l 'op-node/*'` to list tags and
`git fetch upstream --tags`. Give the user a clear list of recent op-node release tags to choose
from. Verify the choice with the user.

**Constrain the choice by the op-geth fork (critical).** The base must line up with the op-geth
fork (see Conventions). Check both:

```bash
# what op-geth version does a candidate op-node tag pin?
git show op-node/v1.17.0:go.mod | grep 'ethereum-optimism/op-geth'
# what upstream op-geth is the celo op-geth fork rebased onto? (its -base/-upstream marker)
#   in the op-geth checkout: git describe --tags celo-rebase-XX-base
```

Choose the op-node tag whose op-geth pin matches the op-geth fork's base. (Example from
celo-rebase-18: op-geth fork was on op-geth `v1.101702.1`, which matches `op-node/v1.17.0`'s pin
`v1.101702.1-rc.1` — **not** v1.18.x, which pin newer op-geth.) Also confirm the chosen tag is a
forward move: `git merge-base --is-ancestor <old-upstream> $BASE`.

Save it for the session:

```bash
export BASE=op-node/v1.17.0
git fetch upstream tag $BASE
```

Create the working branch (do NOT use a `celo-` name yet — it is protected):

```bash
git switch -c user/celo-rebase-XX   # e.g. user/celo-rebase-18
```

Before starting, get a clear understanding of the commits that are added by moving the upstream base to `$BASE`.

Create a list of the commits and their purpose.

Also start a log file called `rebase-XX.log` to track the rebase progress. It should contain the base tag.

## 3. Rebase in passes

Getting a perfect result in one go is hard. First pass: resolve conflicts and make the Go
linter pass after every commit.

```bash
# optimism: lint every commit during the rebase
git rebase $BASE -i --exec 'just lint-go'
```

`just lint-go` builds and runs the repo's own linter (`./linter/bin/op-golangci-lint` via
`cd linter && just build`) — no separate `golangci-lint` install is needed. (`make lint-go`
still works but is deprecated and just forwards to `just lint-go`.)

Two practical notes:
- **Go toolchain must be consistent.** `just lint-go`/builds fail cryptically (`compile: version
  ... does not match go tool version`) if `go` and `GOROOT` disagree. Use the repo-pinned go (via
  mise) and make sure no stale `GOROOT`/`go` from elsewhere shadows it. `go version` and
  `go env GOROOT GOVERSION` must agree.
- **Per-commit lint inflates the Go build cache and can fill the disk** (it grew 8 GB → 27 GB in
  one rebase, causing `no space left on device` mid-lint — which looks like a fake "typecheck"
  failure). Watch disk; `go clean -cache` to reclaim.

**Resolving conflicts.** Use diff3-style markers so you can see the common ancestor (and so the
one-liners below work): `git config merge.conflictStyle zdiff3`. Markers are then
`<<<<<<< HEAD` / `||||||| base` / `=======` / `>>>>>>>` — the middle group tells you who changed
what.
[mergiraf](https://codeberg.org/mergiraf/mergiraf) (`mergiraf solve path/to/file`) occasionally
helps for large structured files, but in practice it usually can't resolve these (YAML, Go struct
literals) — don't spend long on it, resolve manually. Common moves:

- **Union** (both sides added independent things, e.g. imports/struct fields): keep both.
- **Take upstream's side for all conflicts in a file** (the workhorse for CI files & generated
  files we'll regenerate — see §4):
  ```bash
  perl -0777 -pi -e 's/^<<<<<<< .*?\n(.*?)^\|\|\|\|\|\|\| .*?\n.*?^=======\n.*?^>>>>>>> .*?\n/$1/gms' FILE
  ```
- **Take celo's (theirs) side for all conflicts in a file** (capture the middle group instead):
  ```bash
  perl -0777 -pi -e 's/^<<<<<<< .*?\n.*?^\|\|\|\|\|\|\| .*?\n.*?^=======\n(.*?)^>>>>>>> .*?\n/$1/gms' FILE
  ```
- **modify/delete** (`DU`/`UD`, 0 conflict markers): upstream deleted a file celo modified (or
  vice versa). Usually follow upstream — `git rm FILE` (the celo edit, e.g. a test skip, is moot
  if upstream removed the file). Keep it only if it's a celo-owned file.

If a resolution isn't obvious, display a short summary and ask for decisions. **Log every
non-trivial conflict** (one line: commit, file, what each side did, how you resolved) in
`rebase-XX.log`.

Once conflicts and lint are clean, get the tests green. Then look at the history again for any
**safe, non-history-rewriting** cleanups. Larger cleanups should be done as **normal PRs after**
the new rebase is published as the default branch — that is safer and allows proper review.

### Verify the whole branch before publishing

The per-commit `just lint-go` only covers Go. Once the rebase is done (and after the repo-specific
tasks in §4), run the full suite and confirm each is green:

```bash
just build-go                                  # Go binaries + cannon
just lint-go                                   # golangci-lint + go mod tidy -diff
(cd packages/contracts-bedrock && just build)  # forge build (Solidity)
(cd rust && just build && just lint)           # cargo build + clippy/docs for the whole workspace
```

(The rust build/lint and the cold forge build each take ~15–20 min.) Then run tests (see
"Bisect a broken test").

## 4. Repo-specific tasks

### op-geth (do this FIRST — see Conventions)

op-geth is rebased separately in `celo-org/op-geth` as `celo-rebase-XX` (+ a `-upstream`/`-base`
marker), onto the upstream op-geth that the chosen optimism `$BASE` pins. Then point optimism's
go.mod at it:

```bash
# get the pseudo-version for an op-geth commit, then update the replace + tidy:
go list -m github.com/celo-org/op-geth@<commit>
ops/scripts/celo-update-op-geth.py <pseudo-version>   # = go mod edit -replace ... + go mod tidy
```

- The stack usually has several `chore: update op-geth ...` commits pinning *old* op-geth
  pseudo-versions. Once go.mod points at the new fork they are stale — **drop/collapse them**
  (`git rebase --skip` when they conflict in go.mod/go.sum only).
- op-geth's API may differ from upstream go-ethereum (e.g. celo's `core.ApplyTransaction` takes an
  extra `feeCurrencyContext`). When a `gomod: Update op-geth` commit conflicts in Go files, **read
  the real signatures from the op-geth checkout** (don't guess) and merge upstream's new call shape
  with celo's extra args. `go.sum` conflicts: take HEAD, then `go mod tidy` fixes it.
- Celo loads chain/fork config from its **superchain-registry fork, embedded in op-geth**. Newer
  op-geth may drop hardcoded celo constants (e.g. `params.CeloMainnetIsthmusTimestamp`) in favor of
  registry config — and a later optimism commit (`Use config from celo's fork of superchain
  registry`) then removes the op-node shortcut that used them (`applyCeloHardforks`). We do **not**
  rebase superchain-registry itself.

### op-core/superchain (NEW in v1.18.x — its pin must match op-geth)

Upstream extracted the superchain-registry config *out* of op-geth into the new `op-core/superchain`
package (the op-geth decoupling). Its `init()` runs `VerifyEmbeddedCommit()`, which asserts
`op-core/superchain/superchain-registry-commit.txt` **==** op-geth's `EmbeddedRegistryCommit()`. So
for celo it must point at celo's registry **fork at the same commit op-geth embeds**, or
`go test ./op-core/superchain/...` (`TestSyncSuperchain`) fails — and any binary importing the
package panics at startup:

```bash
# 1. op-geth's embedded SR commit (the celo fork):
git -C <op-geth> show celo-rebase-XX:superchain-registry-commit.txt
# 2. write it into op-core/superchain/superchain-registry-commit.txt, and change the clone URL in
#    op-core/superchain/sync-superchain.sh to https://github.com/celo-org/superchain-registry.git
# 3. rebuild the bundle (zip is gitignored; commit the .sha256 + the two edited files):
rm op-core/superchain/superchain-configs.zip && just sync-superchain
```

op-node still reads chain/fork config from op-geth's `go-ethereum/superchain` package during the
transition — `op-core/superchain` is not yet a config source for celo, but its consistency check
must still pass.

### Contract snapshots & semver lock

Do not merge conflicts in generated files — drop the old snapshot/semver commits and recreate:

```bash
cd packages/contracts-bedrock/
just snapshots && just semver-lock
# if it fails, clean and retry:
just clean && forge clean
```

Also regenerate the **NUT (network-upgrade-transaction) upgrade bundle** — a separate snapshot
(`snapshots/upgrades/current-upgrade-bundle.json`) that embeds predeploy implementation bytecode,
so it goes stale whenever celo's contract bytecode changes. The `nut-bundle` check in
`contracts-bedrock-checks-fast-feature-tests` runs `just nut-bundle-check` (regenerate + `git diff
--exit-code`), so a stale snapshot fails CI:

```bash
cd packages/contracts-bedrock/
just generate-nut-bundle
```

`test-deploy-config-full.json` is parsed strictly (`DisallowUnknownFields`): when resolving its
conflicts it must match `op-chain-ops/genesis/config.go` exactly — drop fields upstream removed,
add celo fields.

### rust / Cargo.lock

Treat `rust/Cargo.lock` like a generated file: don't hand-merge it. Take HEAD (`--ours`) and
regenerate after the rebase (`cargo update -p <pkg>` for a specific bump, or let `cargo build`
update it), then verify `cargo metadata --locked` and `cargo deny`.

### CircleCI / GitHub workflows

Upstream restructures CI heavily, so most `.circleci/*` and `.github/workflows/*` commits conflict.
Strategy: keep config **valid now**, do a dedicated CI reconciliation PR **after** the rebase
lands. For each CI conflict, **check whether celo's change can be applied to the upstream-modified
workflow** and port the intent where feasible:

- celo disabled/removed a workflow upstream still has → re-apply the disable (`when: false` / delete)
  in the new structure.
- celo skipped a job upstream renamed/moved → comment it out in the new location **and** comment out
  its entries in any `requires:` / required-gate fan-ins (a `requires:` on a non-scheduled job is an
  invalid CircleCI config).
- upstream deleted a workflow celo also wanted gone → keep it deleted (`git rm`).
- if porting cleanly is hard (deep restructure), take upstream for that conflict and note it for the
  follow-up PR.

Validate after editing: `circleci config validate` and `actionlint`/`yamllint` if installed,
otherwise at least parse the YAML (`ruby -ryaml -e 'YAML.load_file(ARGV[0])' FILE`).

## Avoiding and fixing breakage when editing history

Editing history easily breaks things. Detect and fix problems fast — it is easy to fix one
broken commit, hard to untangle many at once.

### Lint every commit

If a commit breaks the linter, **do not** add a fixup commit on top. Fix it in the commit that
broke it. The `--exec` flag runs the linter after each commit and stops at the first failure:

```bash
git rebase $BASE -i --exec 'just lint-go'
```

When the per-commit `--exec` fails the commit is already made: fix the files, `git add`, then
`git commit --amend --no-edit` and `git rebase --continue` (it does not re-run the failed exec; the
next commit's exec re-checks the whole tree).

Don't run `forge build`/`cargo build` per-commit — they're far too slow (~15–20 min). Build
contracts/rust once in the verify pass (§3). For workflow-only commits, a quick `actionlint` /
`yamllint` / YAML parse is enough.

### Bisect a broken test

Running all tests is slow. After a batch of changes, run the full suite, then bisect to find
the commit that broke a test:

```bash
git bisect start HEAD $BASE
# Go test (add `-count 5` for flaky tests):
git bisect run gotestsum ./op-e2e/system/altda -run TestBatcher_FailoverToEthDA_FallbackToAltDA
# Solidity test:
git bisect run forge test --root packages/contracts-bedrock/ --match-test test_cannotReinitialize_succeeds
```

If a test does not exist in every commit in the range, skip those commits with exit code 125:

```bash
#!/bin/bash
# bisect_test.sh — usage: git bisect run ./bisect_test.sh ./cmd/evm/... -run TestEofParse
output=$(gotestsum "$@" 2>&1)
exit_code=$?
if [[ ! $output =~ "tests" ]]; then
    echo "no tests run"; exit 125   # tell bisect to skip this commit
fi
echo "$output"
exit $exit_code
```

Check that tests merely **compile** (filter for a non-existent test name): `go test -run XXXXXX ./...`.

## Tips

- **Split a commit**: mark it `edit` in the todo list. When the rebase stops, the commit is at
  `HEAD`. Preserve author/date, then re-commit in chunks:
  ```bash
  orig=<commit>
  author="$(git show -s --format='%an <%ae>' "$orig")"
  date="$(git show -s --format='%ad' "$orig")"
  git reset HEAD^                       # unstage the commit's changes, keep them in the tree
  git add -p                            # stage just the first chunk
  git commit --author="$author" --date="$date" -m "first part"
  # repeat stage → commit for each chunk, then:
  git rebase --continue
  ```
