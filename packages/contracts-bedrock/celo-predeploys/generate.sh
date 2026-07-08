#!/usr/bin/env bash
# Regenerate predeploys.json: a single self-contained activation bundle for the
# Celo CGT v2 L2 predeploys (code + storage + balances + params).
#
# Bytecode/codeHash come from forge-artifacts (deterministic). Business storage, balances,
# params, and notes are the curated init-equivalent state (from simulation) embedded below.
# Structural EIP-1967 storage (admin, implementation) is derived per predeploy.
#
# Run from packages/contracts-bedrock after `forge build`. Requires `cast` + python3.
set -euo pipefail
cd "$(dirname "$0")/.."

OUT="celo-predeploys/predeploys.json"

# contract | source | kind | proxyAddress | implementationAddress | solc | evmVersion
read -r -d '' TABLE <<'EOF' || true
L1BlockCGT|src/L2/L1BlockCGT.sol|upgrade|0x4200000000000000000000000000000000000015|0xc0d3C0D3C0D3c0D3C0D3C0d3C0D3c0D3c0d30015|0.8.15+commit.e14f2714|london
L2ToL1MessagePasserCGT|src/L2/L2ToL1MessagePasserCGT.sol|upgrade|0x4200000000000000000000000000000000000016|0xC0D3C0d3C0d3c0d3C0d3C0D3c0D3c0d3c0D30016|0.8.15+commit.e14f2714|london
NativeAssetLiquidity|src/L2/NativeAssetLiquidity.sol|new|0x4200000000000000000000000000000000000029|0xC0d3c0D3C0d3c0D3C0d3c0D3C0d3c0D3C0d30029|0.8.15+commit.e14f2714|london
LiquidityController|src/L2/LiquidityController.sol|new|0x420000000000000000000000000000000000002a|0xC0D3C0d3C0D3C0d3c0D3c0D3C0D3C0D3C0D3002A|0.8.15+commit.e14f2714|london
CeloGasBridgeL2|src/celo/CeloGasBridgeL2.sol|new|0x4200000000000000000000000000000000001023|0xc0d3C0d3C0d3C0D3C0D3c0d3c0d3C0d3C0d31023|0.8.15+commit.e14f2714|london
CeloSequencerFeeVault|src/L2/CeloSequencerFeeVault.sol|upgrade|0x4200000000000000000000000000000000000011|0xC0D3C0d3c0d3c0d3C0D3c0d3C0D3c0d3c0D30011|0.8.25+commit.b61c2a91|cancun
EOF

# Curated init-equivalent state: params, notes, and per-predeploy business storage/balance.
# EIP-1967 admin/impl slots are NOT listed here; the generator derives them.
read -r -d '' STATIC <<'JSON' || true
{
  "params": {
    "liquidityControllerOwner": {
      "note": "Celo Governance",
      "mainnet": "0x000000000000000000000000d533ca259b330c7a88f74e000a3faea2d63b7972",
      "sepolia": "0x00000000000000000000000050d2f15ccf5e97999bdf9d6760d0208b00d14ad1",
      "chaos": "0x000000000000000000000000d98fe96ad096a650a8d30f8bbe73897fe1780edb"
    },
    "celoTokenL1": {
      "note": "L1 CELO ERC20",
      "mainnet": "0x000000000000000000000000057898f3c43f129a17517b9056d23851f124b19f",
      "sepolia": "0x0000000000000000000000003c7011fd5e6aed460caa4985cf8d8caba435b092",
      "chaos": "0x000000000000000000000000894afac3694c29b506d012e32b3d07aaa027a854"
    },
    "celoGasBridgeL1": "<L1 deploy output: CeloGasBridgeL1 proxy, 32-byte left-padded>",
    "nativeAssetLiquidityAmount": "<migration-time: L2 native balance backing bridged CELO, per network>"
  },
  "notes": [
    "Every predeploy address holds the identical OP Proxy shell (proxyShell.bytecode); its EIP-1967 storage points admin -> ProxyAdmin (0x42..0018) and implementation -> impl.address. kind=new plants the shell and writes both slots; kind=upgrade already has the shell + admin on-chain, so only the implementation slot is repointed.",
    "L1BlockCGT repoints its impl slot from the live 0x3ba4..df2c to the namespace 0xc0d3..0015.",
    "celoGasBridgeL1 and nativeAssetLiquidityAmount are filled at migration time; all other values are final."
  ],
  "business": {
    "0x4200000000000000000000000000000000000015": {
      "storage": [
        { "slot": "0xd2ff82c9b477ff6a09f530b1c627ffb4b0b81e2ae2ba427f824162e8dad020aa", "value": "0x0000000000000000000000000000000000000000000000000000000000000001", "note": "isCustomGasToken = true" }
      ]
    },
    "0x4200000000000000000000000000000000000029": {
      "balance": "param:nativeAssetLiquidityAmount",
      "storage": []
    },
    "0x420000000000000000000000000000000000002a": {
      "storage": [
        { "slot": "0x0000000000000000000000000000000000000000000000000000000000000000", "value": "0x0000000000000000000000000000000000000000000000000000000000000001", "note": "_initialized = 1" },
        { "slot": "0x0000000000000000000000000000000000000000000000000000000000000033", "value": "param:liquidityControllerOwner", "note": "_owner (slot 51)" },
        { "slot": "0x0000000000000000000000000000000000000000000000000000000000000066", "value": "0x43656c6f00000000000000000000000000000000000000000000000000000008", "note": "gasPayingTokenName = \"Celo\"" },
        { "slot": "0x0000000000000000000000000000000000000000000000000000000000000067", "value": "0x43454c4f00000000000000000000000000000000000000000000000000000008", "note": "gasPayingTokenSymbol = \"CELO\"" },
        { "slot": "0xde7e0d28ff6c7384356f2757f7714325aba6baaf05019dbc32fc04e101dfee4f", "value": "0x0000000000000000000000000000000000000000000000000000000000000001", "note": "minters[CeloGasBridgeL2 0x42..1023] = true" }
      ]
    },
    "0x4200000000000000000000000000000000001023": {
      "storage": [
        { "slot": "0x0000000000000000000000000000000000000000000000000000000000000000", "value": "0x0000000000000000000000000000000000000000000000000000000000000001", "note": "_initialized = 1" },
        { "slot": "0x0000000000000000000000000000000000000000000000000000000000000003", "value": "0x0000000000000000000000004200000000000000000000000000000000000007", "note": "messenger = L2CrossDomainMessenger" },
        { "slot": "0x0000000000000000000000000000000000000000000000000000000000000004", "value": "param:celoGasBridgeL1", "note": "otherBridge" },
        { "slot": "0x0000000000000000000000000000000000000000000000000000000000000032", "value": "param:celoTokenL1", "note": "celoTokenL1 (slot 50)" }
      ]
    }
  }
}
JSON

artifact_code() { python3 -c "import json;print(json.load(open('forge-artifacts/$1.sol/$1.json'))['deployedBytecode']['object'])"; }

GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
GIT_COMMIT=$(git rev-parse HEAD)
FORGE_VER=$(forge --version | head -1)

{
  echo "$TABLE"
  echo "---PROXY---"
  PROXY_CODE=$(artifact_code Proxy)
  echo "src/universal/Proxy.sol|$PROXY_CODE|$(cast keccak "$PROXY_CODE")"
  echo "---IMPL---"
  while IFS='|' read -r name src kind proxy impl solc evm; do
    [ -z "$name" ] && continue
    code=$(artifact_code "$name")
    echo "$name|$src|$kind|$proxy|$impl|$solc|$evm|$code|$(cast keccak "$code")"
  done <<< "$TABLE"
} | GIT_BRANCH="$GIT_BRANCH" GIT_COMMIT="$GIT_COMMIT" FORGE_VER="$FORGE_VER" STATIC="$STATIC" python3 -c '
import sys, os, json
lines = sys.stdin.read().splitlines()
i = lines.index("---PROXY---")
j = lines.index("---IMPL---")
proxy_src, proxy_code, proxy_hash = lines[i+1:j][0].split("|")
rows = [l.split("|") for l in lines[j+1:] if l.strip()]
static = json.loads(os.environ["STATIC"])
biz = static["business"]

ADMIN_SLOT = "0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103"
IMPL_SLOT  = "0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc"
PROXY_ADMIN = "0x4200000000000000000000000000000000000018"

def word(addr):
    return "0x" + "0" * 24 + addr[2:].lower()

predeploys = []
for name, src, kind, proxy_addr, impl_addr, solc, evm, code, codehash in rows:
    b = biz.get(proxy_addr, {})
    storage = list(b.get("storage", []))
    if kind == "new":
        storage.append({"slot": ADMIN_SLOT, "value": word(PROXY_ADMIN), "note": "EIP-1967 admin slot -> ProxyAdmin (0x42..0018)"})
    storage.append({"slot": IMPL_SLOT, "value": word(impl_addr), "note": "EIP-1967 implementation slot -> impl.address"})
    entry = {
        "name": name,
        "kind": kind,
        "address": proxy_addr,
        "impl": {
            "address": impl_addr,
            "source": src,
            "solc": solc,
            "evmVersion": evm,
            "codeHash": codehash,
            "codeLength": (len(code) - 2) // 2,
            "bytecode": code,
        },
    }
    if "balance" in b:
        entry["balance"] = b["balance"]
    entry["storage"] = storage
    predeploys.append(entry)

out = {
    "description": (
        "One-shot activation state for the Celo CGT v2 L2 predeploys: a single self-contained bundle "
        "(code + storage + balances). For each predeploys[] entry: (1) plant impl.bytecode at impl.address; "
        "(2) ensure the proxy account at address holds proxyShell.bytecode (identical for all predeploys; "
        "already on-chain for kind=upgrade); (3) apply storage (and balance) to the proxy account. A value of "
        "param:X is network-specific -- substitute params.X[network]; every other value is literal and identical "
        "on all networks. eip1967Slots decodes the two structural slots. Verify planted code with codeHash. "
        "Regenerate with generate.sh."
    ),
    "build": {
        "gitBranch": os.environ["GIT_BRANCH"],
        "gitCommit": os.environ["GIT_COMMIT"],
        "forge": os.environ["FORGE_VER"],
        "settings": "optimizer=true, runs=999999, bytecode_hash=none; solc + evm_version are per-contract (0.8.15 -> london, 0.8.25 -> cancun)",
        "derivation": "storage/balances captured by simulating the real init calls and recording the diff, cross-checked with the compiler storage layout, cast index, and live Celo mainnet/sepolia/chaos",
    },
    "eip1967Slots": {"admin": ADMIN_SLOT, "implementation": IMPL_SLOT},
    "proxyShell": {
        "source": proxy_src,
        "codeHash": proxy_hash,
        "codeLength": (len(proxy_code) - 2) // 2,
        "bytecode": proxy_code,
    },
    "params": static["params"],
    "predeploys": predeploys,
}
json.dump(out, open("'"$OUT"'", "w"), indent=2)
print("wrote '"$OUT"' ->", len(predeploys), "predeploys")
'
