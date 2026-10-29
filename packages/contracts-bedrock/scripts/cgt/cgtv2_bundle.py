#!/usr/bin/env python3
"""Generate a CGTv2 bundle from pinned L2 getters, or verify its settings before migration."""
import argparse
import hashlib
import json
import re
from pathlib import Path
import subprocess
import sys
import tempfile
import urllib.error
import urllib.request

ROOT = Path(__file__).resolve().parents[2]
VAULT = "0x4200000000000000000000000000000000000011"
ADDRESSES = ("bridgeProxy", "l1Bridge", "l1Token", "owner")
METADATA = ("gasPayingTokenName", "gasPayingTokenSymbol")
GETTERS = {
    "vaultRecipient": (VAULT, "RECIPIENT()", "address"),
    "vaultMinWithdrawalAmount": (VAULT, "MIN_WITHDRAWAL_AMOUNT()", "uint256"),
    "vaultWithdrawalNetwork": (VAULT, "WITHDRAWAL_NETWORK()", "uint8"),
}


class BundleError(Exception):
    pass


def tool(args):
    try:
        result = subprocess.run(args, cwd=ROOT, text=True, capture_output=True)
    except FileNotFoundError as error:
        raise BundleError(f"{args[0]} is required on PATH") from error
    if result.returncode:
        raise BundleError(f"{args[0]} failed: {result.stderr[-3000:]}")
    return result.stdout


class Rpc:
    def __init__(self, url):
        self.url = url
        self.request_id = 0

    def request(self, method, params):
        self.request_id += 1
        payload = json.dumps({"jsonrpc": "2.0", "id": self.request_id,
                              "method": method, "params": params}).encode()
        request = urllib.request.Request(self.url, payload, {"Content-Type": "application/json"})
        try:
            with urllib.request.urlopen(request, timeout=30) as response:
                data = json.load(response)
        except (urllib.error.URLError, ValueError) as error:
            raise BundleError(f"RPC {method} request failed") from error
        if not isinstance(data, dict):
            raise BundleError(f"RPC {method} returned an invalid response")
        if data.get("id") != self.request_id or "error" in data or "result" not in data:
            raise BundleError(f"RPC {method} failed: {data.get('error', 'invalid response')}")
        return data["result"]


def deployment_config(data):
    if set(data) != {"chainId", *ADDRESSES, *METADATA}:
        raise BundleError("config must contain chainId, bridgeProxy, l1Bridge, l1Token, owner, gasPayingTokenName and gasPayingTokenSymbol only")
    if type(data["chainId"]) is not int or data["chainId"] <= 0:
        raise BundleError("invalid chainId")
    for key in ADDRESSES:
        if not isinstance(data[key], str) or not re.fullmatch(r"0x[0-9a-fA-F]{40}", data[key]):
            raise BundleError(f"invalid {key} address")
    for key in METADATA:
        if not isinstance(data[key], str) or not data[key].strip():
            raise BundleError(f"invalid {key}: nonempty token metadata is required")
        data[key].encode("utf-8")
    return {key: value.lower() if key in ADDRESSES else value for key, value in data.items()}


def read_settings(rpc, expected_chain_id, block_hash=None):
    chain_id = int(rpc.request("eth_chainId", []), 16)
    if chain_id != expected_chain_id:
        raise BundleError(f"RPC chain mismatch: expected {expected_chain_id}, got {chain_id}")
    if block_hash is None:
        block = rpc.request("eth_getBlockByNumber", ["latest", False])
        if (not isinstance(block, dict) or not isinstance(block.get("hash"), str)
                or not re.fullmatch(r"0x[0-9a-fA-F]{64}", block["hash"])):
            raise BundleError("RPC latest block is missing or has an invalid block hash")
        block_hash = block["hash"].lower()
    else:
        if not re.fullmatch(r"0x[0-9a-fA-F]{64}", block_hash):
            raise BundleError("--block-hash must be an explicit block hash; omit it to use latest")
        block_hash = block_hash.lower()
        block = rpc.request("eth_getBlockByHash", [block_hash, False])
    if not block or block.get("hash", "").lower() != block_hash:
        raise BundleError("RPC block hash mismatch or missing block")
    number = int(block["number"], 16)
    settings = {}
    for field, (address, signature, return_type) in GETTERS.items():
        selector = tool(["cast", "sig", signature]).strip()
        raw = rpc.request("eth_call", [{"to": address, "data": selector},
                                     {"blockHash": block_hash, "requireCanonical": True}])
        try:
            value = json.loads(tool(["cast", "abi-decode", f"f()({return_type})", raw, "--json"]))[0]
            if return_type == "address":
                value = value.lower()
            elif return_type.startswith("uint"):
                value = int(value)
        except (BundleError, ValueError, IndexError, TypeError) as error:
            raise BundleError(f"invalid {signature} result") from error
        if field == "vaultWithdrawalNetwork" and value not in (0, 1):
            raise BundleError("invalid WITHDRAWAL_NETWORK() result")
        settings[field] = value
    return block_hash, number, settings


def snapshot_path(bundle):
    return bundle.with_suffix(".snapshot.json")


def generate(args):
    config = deployment_config(json.loads(Path(args.config).read_text()))
    block_hash, number, settings = read_settings(Rpc(args.rpc_url), config["chainId"], args.block_hash)
    settings |= {key: config[key] for key in METADATA}
    network = {key: config[key] for key in ADDRESSES} | settings
    bundle = Path(args.output).resolve()
    bundle.parent.mkdir(parents=True, exist_ok=True)
    scratch = ROOT / ".testdata"
    scratch.mkdir(exist_ok=True)
    # JSON avoids Forge CLI tuple parsing of quotes, backslashes and Unicode metadata.
    with tempfile.TemporaryDirectory(prefix="cgtv2-settings-", dir=scratch) as directory:
        captured = Path(directory) / "settings.json"
        captured.write_text(json.dumps(network | {
            "vaultMinWithdrawalAmount": str(settings["vaultMinWithdrawalAmount"])
        }))
        tool(["forge", "script", "--offline", "scripts/cgt/GenerateCgtV2NUTBundle.s.sol",
              "--sig", "runProductionFromJson(string,string)", str(captured), str(bundle)])
    if bundle_settings(bundle) != settings:
        raise BundleError("generated initializer calldata does not match the configured metadata and captured vault settings")
    record = {"chainId": config["chainId"], "blockHash": block_hash,
              "blockNumber": number, "networkConfig": network,
              "bundleSha256": hashlib.sha256(bundle.read_bytes()).hexdigest()}
    snapshot_path(bundle).write_text(json.dumps(record, indent=2) + "\n")
    print(f"Wrote {bundle} and {snapshot_path(bundle)} from L2 block {record['blockHash']}")


def bundle_settings(bundle):
    transactions = json.loads(bundle.read_text())["transactions"]
    settings = {}
    # Decode the initializer calls in the frozen direct-upgrade group itself.
    for step, signature, fields in [
        (9, "f()(address,bytes,bytes)",
         (None, "gasPayingTokenName", "gasPayingTokenSymbol")),
        (11, "f()(address,uint256,uint8)",
         ("vaultRecipient", "vaultMinWithdrawalAmount", "vaultWithdrawalNetwork")),
    ]:
        outer = json.loads(tool(["cast", "calldata-decode", "upgradeToAndCall(address,bytes)",
                                 transactions[step]["data"], "--json"]))
        values = json.loads(tool(["cast", "abi-decode", signature, "0x" + outer[1][10:], "--json"]))
        for field, value in zip(fields, values, strict=True):
            if field is not None:
                if field in ("vaultMinWithdrawalAmount", "vaultWithdrawalNetwork"):
                    value = int(value)
                elif field in ("gasPayingTokenName", "gasPayingTokenSymbol"):
                    value = bytes.fromhex(value[2:]).decode("utf-8")
                else:
                    value = value.lower()
                settings[field] = value
    return settings


def verify(args):
    bundle = Path(args.bundle)
    record = json.loads(Path(args.snapshot or snapshot_path(bundle)).read_text())
    if hashlib.sha256(bundle.read_bytes()).hexdigest() != record["bundleSha256"]:
        raise BundleError("bundle does not match the recorded SHA-256")
    expected = bundle_settings(bundle)
    if any(expected[field] != record["networkConfig"][field] for field in (*GETTERS, *METADATA)):
        raise BundleError("snapshot settings do not match the bundle initializer calldata")
    block_hash, number, actual = read_settings(Rpc(args.rpc_url), record["chainId"], args.block_hash)
    if number < record["blockNumber"]:
        raise BundleError("verification block predates the generation snapshot")
    changed = [field for field in GETTERS if actual[field] != expected[field]]
    if changed:
        raise BundleError("legacy settings changed: " + ", ".join(changed))
    print(f"Legacy settings match at L2 block {block_hash} ({number})")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    subparsers = parser.add_subparsers(dest="command", required=True)
    for command in ("generate", "verify"):
        subparser = subparsers.add_parser(command)
        subparser.add_argument("--rpc-url", required=True)
        subparser.add_argument("--block-hash", help="L2 block hash to replay (default: resolve latest once)")
        if command == "generate":
            subparser.add_argument("--config", required=True)
            subparser.add_argument("--output", required=True)
        else:
            subparser.add_argument("--bundle", required=True)
            subparser.add_argument("--snapshot")
    args = parser.parse_args()
    try:
        (generate if args.command == "generate" else verify)(args)
    except (BundleError, OSError, ValueError, KeyError, TypeError) as error:
        print(f"CGTv2 bundle: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
