"""Exercise production bundle generation and preflight through a local JSON-RPC server."""
import hashlib
import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
import subprocess
import sys
import tempfile
import threading
import unittest

ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "scripts/cgt/cgtv2_bundle.py"
BLOCK = "0x" + "11" * 32
NEXT_BLOCK = "0x" + "22" * 32
VAULT = "0x4200000000000000000000000000000000000011"
L1_BLOCK = "0x4200000000000000000000000000000000000015"
DEPLOYMENT = {"chainId": 1337, "bridgeProxy": "0x42000000000000000000000000000000000007fe",
              "l1Bridge": "0x00000000000000000000000000000000000000cc",
              "l1Token": "0x00000000000000000000000000000000000000bb",
              "owner": "0x00000000000000000000000000000000000000aa",
              "gasPayingTokenName": "Celo", "gasPayingTokenSymbol": "CELO"}


def word(value):
    return "0x" + f"{value:064x}"


def string_result(value):
    raw = value.encode()
    return "0x" + f"{32:064x}{len(raw):064x}" + raw.hex().ljust(((len(raw) + 31) // 32) * 64, "0")


class BundleWorkflowTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        (ROOT / ".testdata").mkdir(exist_ok=True)
        cls.temp = tempfile.TemporaryDirectory(prefix="cgtv2-bundle-", dir=ROOT / ".testdata")
        cls.directory = Path(cls.temp.name)
        cls.base = None
        cls.calls = []
        cls.values = {}
        cls.error = None
        cls.chain = "0x539"
        cls.bad_header = False
        cls.latest_block = {"hash": BLOCK, "number": "0x64"}

        class Handler(BaseHTTPRequestHandler):
            def log_message(self, *args):
                pass

            def do_POST(self):
                request = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
                cls.calls.append(request)
                method, params = request["method"], request["params"]
                response = {"jsonrpc": "2.0", "id": request["id"]}
                if method == "eth_chainId":
                    response["result"] = cls.chain
                elif method == "eth_getBlockByNumber":
                    if params != ["latest", False]:
                        response["error"] = {"code": -32602, "message": "unexpected block selector"}
                    else:
                        response["result"] = cls.latest_block
                        # Advance the head immediately, so subsequent calls must keep the captured hash.
                        cls.latest_block = {"hash": NEXT_BLOCK, "number": "0x65"}
                elif method == "eth_getBlockByHash":
                    response["result"] = {"hash": ("0x" + "ff" * 32) if cls.bad_header else params[0],
                                          "number": "0x64" if params[0] == BLOCK else "0x65"}
                elif method == "eth_call":
                    call, block = params
                    selector = call["data"]
                    address = VAULT if selector in ("0x0d9019e1", "0xd3e5792b", "0xd0e12f90") else L1_BLOCK
                    if block not in ({"blockHash": BLOCK, "requireCanonical": True},
                                     {"blockHash": NEXT_BLOCK, "requireCanonical": True}) or call["to"].lower() != address:
                        response["error"] = {"code": -32602, "message": "unpinned or wrong getter"}
                    elif cls.error:
                        response["error"] = cls.error
                    elif selector in cls.values:
                        response["result"] = cls.values[selector]
                    else:
                        response["error"] = {"code": -32602, "message": "unexpected getter"}
                else:
                    response["error"] = {"code": -32601, "message": "unexpected method"}
                data = json.dumps(response).encode()
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(data)))
                self.end_headers()
                self.wfile.write(data)

        cls.server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        cls.thread = threading.Thread(target=cls.server.serve_forever, daemon=True)
        cls.thread.start()
        cls.rpc = f"http://127.0.0.1:{cls.server.server_port}"

    @classmethod
    def tearDownClass(cls):
        cls.server.shutdown()
        cls.server.server_close()
        cls.thread.join()
        cls.temp.cleanup()

    def setUp(self):
        type(self).calls = []
        type(self).error = None
        type(self).chain = "0x539"
        type(self).bad_header = False
        type(self).latest_block = {"hash": BLOCK, "number": "0x64"}
        type(self).values = {"0x0d9019e1": word(0xdd), "0xd3e5792b": word(23),
                            "0xd0e12f90": word(1), "0xd8444715": string_result("Ether"),
                            "0x550fcdc9": string_result("ETH")}

    def command(self, *args):
        return subprocess.run([sys.executable, str(SCRIPT), *args], cwd=ROOT,
                              text=True, capture_output=True)

    def baseline(self):
        cls = type(self)
        if cls.base is None:
            config = self.directory / "deployment.json"
            config.write_text(json.dumps(DEPLOYMENT))
            bundle = self.directory / "bundle.json"
            result = self.command("generate", "--rpc-url", self.rpc, "--block-hash", BLOCK,
                                  "--config", str(config), "--output", str(bundle))
            self.assertEqual(result.returncode, 0, result.stderr)
            cls.base = bundle
        return cls.base

    def verify(self, block=NEXT_BLOCK, bundle=None):
        original = self.baseline()
        return self.command("verify", "--rpc-url", self.rpc, "--block-hash", block,
                            "--bundle", str(bundle or original),
                            "--snapshot", str(original.with_suffix(".snapshot.json")))

    def test_generate_uses_configured_metadata_not_legacy_l1block(self):
        self.values.update({"0xd8444715": string_result("Ether"),
                            "0x550fcdc9": string_result("ETH")})
        config = self.directory / "explicit-metadata-deployment.json"
        config.write_text(json.dumps(DEPLOYMENT))
        bundle = self.directory / "explicit-metadata-bundle.json"
        result = self.command("generate", "--rpc-url", self.rpc, "--block-hash", BLOCK,
                              "--config", str(config), "--output", str(bundle))
        self.assertEqual(result.returncode, 0, result.stderr)
        snapshot = json.loads(bundle.with_suffix(".snapshot.json").read_text())
        self.assertEqual(snapshot["networkConfig"]["gasPayingTokenName"], "Celo")
        self.assertEqual(snapshot["networkConfig"]["gasPayingTokenSymbol"], "CELO")
        calls = [r for r in self.calls if r["method"] == "eth_call"]
        self.assertEqual(len(calls), 3)
        self.assertTrue(all(r["params"][0]["to"] == VAULT for r in calls))
        result = self.command("verify", "--rpc-url", self.rpc, "--block-hash", NEXT_BLOCK,
                              "--bundle", str(bundle))
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_generate_rejects_missing_or_invalid_metadata(self):
        config = self.directory / "invalid-metadata-deployment.json"
        for key in ("gasPayingTokenName", "gasPayingTokenSymbol"):
            for value in (None, "", "   ", 42, False):
                with self.subTest(key=key, value=value):
                    data = dict(DEPLOYMENT)
                    if value is None:
                        del data[key]
                    else:
                        data[key] = value
                    config.write_text(json.dumps(data))
                    bundle = self.directory / "invalid-metadata-bundle.json"
                    result = self.command("generate", "--rpc-url", self.rpc, "--block-hash", BLOCK,
                                          "--config", str(config), "--output", str(bundle))
                    self.assertNotEqual(result.returncode, 0)
                    self.assertFalse(bundle.exists())
                    self.assertFalse(bundle.with_suffix(".snapshot.json").exists())

    def test_generate_defaults_to_pinned_latest(self):
        config = self.directory / "latest-deployment.json"
        config.write_text(json.dumps(DEPLOYMENT))
        bundle = self.directory / "latest-bundle.json"
        result = self.command("generate", "--rpc-url", self.rpc,
                              "--config", str(config), "--output", str(bundle))
        self.assertEqual(result.returncode, 0, result.stderr)
        snapshot = json.loads(bundle.with_suffix(".snapshot.json").read_text())
        self.assertEqual(snapshot["blockHash"], BLOCK)
        self.assertEqual(snapshot["blockNumber"], 100)
        self.assertIn(BLOCK, result.stdout)
        self.assertEqual(len([r for r in self.calls if r["method"] == "eth_getBlockByNumber"]), 1)
        calls = [r for r in self.calls if r["method"] == "eth_call"]
        self.assertEqual(len(calls), 3)
        self.assertTrue(all(c["params"][1] == {"blockHash": BLOCK, "requireCanonical": True} for c in calls))

    def test_verify_defaults_to_pinned_latest(self):
        bundle = self.baseline()
        original_snapshot = bundle.with_suffix(".snapshot.json").read_bytes()
        type(self).calls = []
        type(self).latest_block = {"hash": NEXT_BLOCK, "number": "0x65"}
        result = self.command("verify", "--rpc-url", self.rpc, "--bundle", str(bundle))
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn(NEXT_BLOCK, result.stdout)
        self.assertEqual(bundle.with_suffix(".snapshot.json").read_bytes(), original_snapshot)
        self.assertEqual(len([r for r in self.calls if r["method"] == "eth_getBlockByNumber"]), 1)
        calls = [r for r in self.calls if r["method"] == "eth_call"]
        self.assertEqual(len(calls), 3)
        self.assertTrue(all(c["params"][1] == {"blockHash": NEXT_BLOCK, "requireCanonical": True} for c in calls))

    def test_generate_rejects_missing_or_invalid_latest_block(self):
        config = self.directory / "invalid-latest-deployment.json"
        config.write_text(json.dumps(DEPLOYMENT))
        bundle = self.directory / "invalid-latest-bundle.json"
        for block in (None, {"hash": "0x1234", "number": "0x64"}):
            with self.subTest(block=block):
                type(self).latest_block = block
                type(self).calls = []
                result = self.command("generate", "--rpc-url", self.rpc,
                                      "--config", str(config), "--output", str(bundle))
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("block", result.stderr.lower())
                self.assertTrue(any(r["method"] == "eth_getBlockByNumber" for r in self.calls))
                self.assertFalse(any(r["method"] == "eth_call" for r in self.calls))
                self.assertFalse(bundle.exists())
                self.assertFalse(bundle.with_suffix(".snapshot.json").exists())

    def test_generate_reads_pinned_getters_into_actual_initializer_calldata(self):
        bundle = self.baseline()
        data = json.loads(bundle.read_text())
        snapshot = json.loads(bundle.with_suffix(".snapshot.json").read_text())
        self.assertEqual(snapshot["blockHash"], BLOCK)
        self.assertEqual(snapshot["blockNumber"], 100)
        self.assertEqual(snapshot["chainId"], 1337)
        self.assertFalse(any(r["method"] == "eth_getBlockByNumber" for r in self.calls))
        self.assertEqual(snapshot["bundleSha256"], hashlib.sha256(bundle.read_bytes()).hexdigest())
        self.assertEqual(snapshot["networkConfig"]["vaultRecipient"], "0x00000000000000000000000000000000000000dd")
        self.assertEqual(snapshot["networkConfig"]["vaultMinWithdrawalAmount"], 23)
        self.assertEqual(snapshot["networkConfig"]["gasPayingTokenName"], "Celo")
        self.assertEqual(len(data["transactions"]), 15)
        for index, signature, expected in [
            (11, "initialize(address,uint256,uint8)", ["0x00000000000000000000000000000000000000dd", 23, 1]),
            (9, "initialize(address,string,string)", ["0xdeaddeaddeaddeaddeaddeaddeaddeaddead0001", "Celo", "CELO"]),
        ]:
            outer = json.loads(subprocess.check_output(["cast", "calldata-decode", "upgradeToAndCall(address,bytes)",
                               data["transactions"][index]["data"], "--json"], text=True))
            inner = json.loads(subprocess.check_output(["cast", "calldata-decode", signature, outer[1],
                               "--json"], text=True))
            if index == 11:
                inner[1:] = [int(value) for value in inner[1:]]
            self.assertEqual([v.lower() if isinstance(v, str) and v.startswith("0x") else v for v in inner], expected)
        calls = [r for r in self.calls if r["method"] == "eth_call"]
        self.assertEqual(len(calls), 3)
        self.assertTrue(all(c["params"][1] == {"blockHash": BLOCK, "requireCanonical": True} for c in calls))

    def test_generate_preserves_nondefault_values_from_rpc(self):
        self.values.update({"0x0d9019e1": word(0x1234), "0xd3e5792b": word(99),
                            "0xd0e12f90": word(0), "0xd8444715": string_result("Fixture Coin"),
                            "0x550fcdc9": string_result("FIX")})
        config = self.directory / "custom-deployment.json"
        config.write_text(json.dumps(DEPLOYMENT | {"gasPayingTokenName": "Fixture Coin",
                                                   "gasPayingTokenSymbol": "FIX"}))
        bundle = self.directory / "custom-bundle.json"
        result = self.command("generate", "--rpc-url", self.rpc, "--block-hash", BLOCK,
                              "--config", str(config), "--output", str(bundle))
        self.assertEqual(result.returncode, 0, result.stderr)
        data = json.loads(bundle.read_text())
        for index, signature, expected in [
            (11, "initialize(address,uint256,uint8)", ["0x0000000000000000000000000000000000001234", 99, 0]),
            (9, "initialize(address,string,string)", ["0xdeaddeaddeaddeaddeaddeaddeaddeaddead0001", "Fixture Coin", "FIX"]),
        ]:
            outer = json.loads(subprocess.check_output(["cast", "calldata-decode", "upgradeToAndCall(address,bytes)",
                               data["transactions"][index]["data"], "--json"], text=True))
            inner = json.loads(subprocess.check_output(["cast", "calldata-decode", signature, outer[1],
                               "--json"], text=True))
            if index == 11:
                inner[1:] = [int(value) for value in inner[1:]]
            self.assertEqual([v.lower() if isinstance(v, str) and v.startswith("0x") else v for v in inner], expected)

    def test_generate_preserves_metadata_bytes(self):
        name, symbol = 'Célo "Coin", \\Token\n', 'CÉLO'
        config = self.directory / "metadata-deployment.json"
        config.write_text(json.dumps(DEPLOYMENT | {"gasPayingTokenName": name,
                                                   "gasPayingTokenSymbol": symbol}))
        bundle = self.directory / "metadata-bundle.json"
        result = self.command("generate", "--rpc-url", self.rpc, "--block-hash", BLOCK,
                              "--config", str(config), "--output", str(bundle))
        self.assertEqual(result.returncode, 0, result.stderr)
        transactions = json.loads(bundle.read_text())["transactions"]
        outer = json.loads(subprocess.check_output([
            "cast", "calldata-decode", "upgradeToAndCall(address,bytes)",
            transactions[9]["data"], "--json"], text=True))
        # Decode strings as bytes so Cast display escaping cannot hide a changed payload.
        inner = json.loads(subprocess.check_output([
            "cast", "abi-decode", "f()(address,bytes,bytes)", "0x" + outer[1][10:], "--json"], text=True))
        self.assertEqual(bytes.fromhex(inner[1][2:]), name.encode())
        self.assertEqual(bytes.fromhex(inner[2][2:]), symbol.encode())
        result = self.command("verify", "--rpc-url", self.rpc, "--block-hash", NEXT_BLOCK,
                              "--bundle", str(bundle))
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_generate_and_verify_preserve_uint256_minimum(self):
        self.values["0xd3e5792b"] = word(2**256 - 1)
        config = self.directory / "uint256-deployment.json"
        config.write_text(json.dumps(DEPLOYMENT))
        bundle = self.directory / "uint256-bundle.json"
        result = self.command("generate", "--rpc-url", self.rpc, "--block-hash", BLOCK,
                              "--config", str(config), "--output", str(bundle))
        self.assertEqual(result.returncode, 0, result.stderr)
        result = self.command("verify", "--rpc-url", self.rpc, "--block-hash", NEXT_BLOCK,
                              "--bundle", str(bundle))
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_generate_rpc_failure_writes_no_bundle_or_snapshot(self):
        config = self.directory / "failed-deployment.json"
        config.write_text(json.dumps(DEPLOYMENT))
        bundle = self.directory / "failed-bundle.json"
        type(self).error = {"code": -32000, "message": "missing trie node"}
        result = self.command("generate", "--rpc-url", self.rpc, "--block-hash", BLOCK,
                              "--config", str(config), "--output", str(bundle))
        self.assertNotEqual(result.returncode, 0)
        self.assertFalse(bundle.exists())
        self.assertFalse(bundle.with_suffix(".snapshot.json").exists())

    def test_verify_accepts_unchanged_settings(self):
        result = self.verify()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn(NEXT_BLOCK, result.stdout)

    def test_verify_rejects_each_changed_setting(self):
        self.baseline()
        for selector, value, field in [
            ("0x0d9019e1", word(0xee), "vaultRecipient"),
            ("0xd3e5792b", word(24), "vaultMinWithdrawalAmount"),
            ("0xd0e12f90", word(0), "vaultWithdrawalNetwork"),
        ]:
            with self.subTest(field=field):
                previous = self.values[selector]
                self.values[selector] = value
                result = self.verify()
                self.assertNotEqual(result.returncode, 0)
                self.assertIn(field, result.stderr)
                self.values[selector] = previous

    def test_verify_rejects_modified_bundle(self):
        original = self.baseline()
        changed = self.directory / "modified.json"
        changed.write_bytes(original.read_bytes() + b" ")
        result = self.verify(bundle=changed)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("bundle", result.stderr.lower())

    def test_verify_rejects_snapshot_settings_not_encoded_in_bundle(self):
        original = self.baseline()
        source = original.with_suffix(".snapshot.json")
        record = json.loads(source.read_text())
        record["networkConfig"]["vaultMinWithdrawalAmount"] = 24
        changed_snapshot = self.directory / "changed.snapshot.json"
        changed_snapshot.write_text(json.dumps(record))
        self.values["0xd3e5792b"] = word(24)
        result = self.command("verify", "--rpc-url", self.rpc, "--block-hash", NEXT_BLOCK,
                              "--bundle", str(original), "--snapshot", str(changed_snapshot))
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("calldata", result.stderr)

    def test_verify_rejects_snapshot_metadata_not_encoded_in_bundle(self):
        original = self.baseline()
        source = original.with_suffix(".snapshot.json")
        for field in ("gasPayingTokenName", "gasPayingTokenSymbol"):
            with self.subTest(field=field):
                record = json.loads(source.read_text())
                record["networkConfig"][field] = "Changed"
                snapshot = self.directory / (field + ".snapshot.json")
                snapshot.write_text(json.dumps(record))
                result = self.command("verify", "--rpc-url", self.rpc, "--block-hash", NEXT_BLOCK,
                                      "--bundle", str(original), "--snapshot", str(snapshot))
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("calldata", result.stderr)

    def test_verify_rejects_wrong_chain(self):
        self.baseline()
        type(self).chain = "0x1"
        result = self.verify()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("chain", result.stderr.lower())

    def test_verify_rejects_rpc_error_without_defaults(self):
        self.baseline()
        for error in [{"code": -32000, "message": "missing trie node"},
                      {"code": -32602, "message": "blockHash unsupported"}]:
            with self.subTest(error=error):
                type(self).error = error
                result = self.verify()
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("RPC", result.stderr)

    def test_verify_rejects_bad_header_and_unpinned_tags(self):
        self.baseline()
        type(self).bad_header = True
        result = self.verify()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("hash", result.stderr.lower())
        result = self.verify(block="latest")
        self.assertNotEqual(result.returncode, 0)

    def test_verify_rejects_malformed_getter_data(self):
        self.baseline()
        self.values["0xd0e12f90"] = "0x"
        result = self.verify()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("WITHDRAWAL_NETWORK", result.stderr)


if __name__ == "__main__":
    unittest.main()
