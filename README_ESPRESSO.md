# Running the Espresso tests

The repo-root `flake.nix` provides the exact toolchain CI uses — go 1.26, just, gotestsum and **forge 1.2.3** (newer forge fails the contracts build). No mise, no docker.

## One-time setup

```sh
nix develop                                   # or prefix any command with: nix develop -c
git submodule update --init --recursive
just build-contracts                          # ~10 min
```

## Espresso e2e tests

```sh
just go-tests-espresso-e2e                                          # whole suite, ~1 h serially
just go-tests-espresso-e2e TestE2eDevnetWithEspressoSimpleTransactions   # one test (Go -run regexp)
```

The recipe runs `espresso/environment`, `espresso/devnet-tests` and `espresso/enclave-tests` exactly as `.github/workflows/espresso-e2e-tests.yaml` does; CI only splits it into shards, whose filters live in that file. It refuses to start if `forge-artifacts` is missing and points you at `just build-contracts`. `devnet-tests` and `enclave-tests` are skipped stubs for now — see their package docs.


## Unit tests

```sh
go test ./op-batcher/batcher/...              # includes the espresso_*_test.go files
```
