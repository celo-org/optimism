#!/usr/bin/env python3
"""Freeze CGTv2 fixture deposits and provenance after GenerateCgtV2NUTBundle.runFixture().

Run from packages/contracts-bedrock. Uses Foundry's cast for Ethereum Keccak.
The profile is synthetic chain 1337; it is not a production bundle.
"""
import hashlib
import json
import subprocess
from pathlib import Path

OUT = Path('snapshots/upgrades')

def keccak(data):
    return bytes.fromhex(subprocess.check_output(['cast', 'keccak', '0x' + data.hex()], text=True).strip()[2:])

def rlp(value):
    if isinstance(value, int):
        value = value.to_bytes((value.bit_length() + 7) // 8, 'big')
    if isinstance(value, list):
        payload = b''.join(map(rlp, value))
        offset = 192
    else:
        if len(value) == 1 and value[0] < 128:
            return value
        payload, offset = value, 128
    if len(payload) < 56:
        return bytes([offset + len(payload)]) + payload
    size = len(payload).to_bytes((len(payload).bit_length() + 7) // 8, 'big')
    return bytes([offset + 55 + len(size)]) + size + payload

def source_hash(intent):
    return keccak((2).to_bytes(32, 'big') + keccak(intent.encode()))

def encode(tx, source, mint=0, value=0):
    return b'\x7e' + rlp([
        source, bytes.fromhex(tx['from'][2:]),
        bytes.fromhex(tx['to'][2:]) if tx['to'] else b'',
        mint, value, tx['gasLimit'], 0, bytes.fromhex(tx['data'][2:]),
    ])

bundle = json.loads((OUT / 'cgtv2-fixture-nut-bundle.json').read_text())
rows = []
for i, tx in enumerate(bundle['transactions']):
    intent = f"Cgtv2 {i}: {tx['intent']}"
    source = source_hash(intent)
    encoded = encode(tx, source)
    (OUT / f'cgtv2-{i:02}.rlp').write_bytes(encoded)
    rows.append({'intent': intent, 'sourceHash': '0x' + source.hex(),
                 'transactionHash': '0x' + keccak(encoded).hex(), 'file': f'cgtv2-{i:02}.rlp'})
funding = {'from': '0x420000000000000000000000000000000000002a',
           'to': '0x4200000000000000000000000000000000000029',
           'gasLimit': 100_000, 'data': '0x' + keccak(b'deposit()')[:4].hex()}
funding_intent = 'Celo: CGTv2 Reserve Funding'
funding_source = source_hash(funding_intent)
funding_bytes = encode(funding, funding_source, mint=400, value=400)
(OUT / 'cgtv2-funding-400.rlp').write_bytes(funding_bytes)
rows.append({'intent': funding_intent, 'sourceHash': '0x' + funding_source.hex(),
             'transactionHash': '0x' + keccak(funding_bytes).hex(), 'file': 'cgtv2-funding-400.rlp',
             'migratedAmount': '600', 'totalSupply': '1000', 'reserve': '400'})
(OUT / 'cgtv2-vectors.json').write_text(json.dumps(rows, indent=2) + '\n')
artifacts = ['StorageSetter', 'L2ProxyAdmin', 'L1BlockCGT', 'L2ToL1MessagePasserCGT',
             'NativeAssetLiquidity', 'LiquidityController', 'CeloGasBridgeL2',
             'CeloSequencerFeeVault', 'CeloCGTV2ContractsManager']
builds = {}
for name in artifacts:
    artifact = json.loads(Path(f'forge-artifacts/{name}.sol/{name}.json').read_text())
    metadata = artifact['metadata']
    if isinstance(metadata, str):
        metadata = json.loads(metadata)
    builds[name] = {'compiler': metadata['compiler'], 'settings': metadata['settings']}
files = sorted(p for p in OUT.glob('cgtv2*') if p.name != 'cgtv2-provenance.json')
provenance = {
    'fixtureOnly': True, 'chainId': 1337,
    'sourceCommit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip(),
    'contractBase': '0296412d673a666af6236efa1f615032b1f1c424',
    'upstreamHelpers': '6c5a77cc4be1148712d36b50a35e1f74cf930314',
    'forgeVersion': subprocess.check_output(['forge', '--version'], text=True).strip(),
    'generation': ['forge script --offline scripts/cgt/GenerateCgtV2NUTBundle.s.sol --sig runFixture()',
                   'python3 scripts/cgt/export-cgtv2-fixture.py'],
    'staticGasLimit': sum(t['gasLimit'] for t in bundle['transactions']),
    'fundingGasLimit': funding['gasLimit'],
    'sha256': {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in files},
    'builds': builds,
    'unresolved': ['production reserved bridge address', 'production network configuration',
                   'production total supply', 'production prestate and deployer nonce checks',
                   'production gas limits'],
}
(OUT / 'cgtv2-provenance.json').write_text(json.dumps(provenance, indent=2) + '\n')
print(f'Exported {len(rows)} deposit vectors; static gas limit {provenance["staticGasLimit"]}')
