#!/usr/bin/env bash
set -euo pipefail

# Baklava: V2
# OPCM=0xd29841fbcff24eb5157f2abe7ed0b9819340159a
# DWI=0x1e121e21e1a11ae47c0efe8a7e13ae3eb4923796
# OPI=0xbed463769920dac19a7e2adf47b6c6bb6480bd97
# SCI=0x911ea44d22eb903515378625da3a0e09d2e1b074
# LCDMI=0x3d5a67747de7e09b0d71f5d782c8b45f6307b9fd
# LEBI=0x276d3730f219f7ec22274f7263180b8452b46d47
# LSBI=0xaf38504abc62f28e419622506698c5fa3ca15eda
# OMEFI=0x5493f4677a186f64805fe7317d6993ba4863988f
# DGFI=0x4bba758f006ef09402ef31724203f316ab74e4a0
# ASRI=0x7b465370bb7a333f99edd19599eb7fb1c2d3f8d2
# SUI=0x4da82a327773965b8d4d85fa3db8249b387458e7
# PVI=0x37e15e4d6dffa9e5e320ee1ec036922e563cb76c

# Baklava: V3
# OPCM=0xdd07cb5e4b2e89a618f8d3a08c8ff753acfe1c68
# DWI=0x1e121e21e1a11ae47c0efe8a7e13ae3eb4923796
# OPI=0x215a5ff85308a72a772f09b520da71d3520e9ac7
# SCI=0x9c61c5a8ff9408b83ac92571278550097a9d2bb5
# LCDMI=0x807124f75ff2120b2f26d7e6f9e39c03ee9de212
# LEBI=0x7ae1d3bd877a4c5ca257404ce26be93a02c98013
# LSBI=0x28841965b26d41304905a836da5c0921da7dbb84
# OMEFI=0x6a52641d87a600ba103ccdfbe3eb02ac7e73c04a
# DGFI=0x4bba758f006ef09402ef31724203f316ab74e4a0
# ASRI=0x7b465370bb7a333f99edd19599eb7fb1c2d3f8d2
# SUI=0x4da82a327773965b8d4d85fa3db8249b387458e7
# PVI=0x37e15e4d6dffa9e5e320ee1ec036922e563cb76c

# Alfajores: V2
# OPCM=0x643e6bcf2708bca3847d30719d94f405c5700c6a
# DWI=0x1e121e21e1a11ae47c0efe8a7e13ae3eb4923796
# OPI=0xbed463769920dac19a7e2adf47b6c6bb6480bd97
# SCI=0x911ea44d22eb903515378625da3a0e09d2e1b074
# LCDMI=0x3d5a67747de7e09b0d71f5d782c8b45f6307b9fd
# LEBI=0x276d3730f219f7ec22274f7263180b8452b46d47
# LSBI=0xaf38504abc62f28e419622506698c5fa3ca15eda
# OMEFI=0x5493f4677a186f64805fe7317d6993ba4863988f
# DGFI=0x4bba758f006ef09402ef31724203f316ab74e4a0
# ASRI=0x7b465370bb7a333f99edd19599eb7fb1c2d3f8d2
# SUI=0x4da82a327773965b8d4d85fa3db8249b387458e7
# PVI=0x37e15e4d6dffa9e5e320ee1ec036922e563cb76c

# Alfajores: V3
# OPCM=0xaeded1bdf59805ed6298da57b0cb974dcc5feb48
# DWI=0x1e121e21e1a11ae47c0efe8a7e13ae3eb4923796
# OPI=0x215a5ff85308a72a772f09b520da71d3520e9ac7
# SCI=0x9c61c5a8ff9408b83ac92571278550097a9d2bb5
# LCDMI=0x807124f75ff2120b2f26d7e6f9e39c03ee9de212
# LEBI=0x7ae1d3bd877a4c5ca257404ce26be93a02c98013
# LSBI=0x28841965b26d41304905a836da5c0921da7dbb84
# OMEFI=0x6a52641d87a600ba103ccdfbe3eb02ac7e73c04a
# DGFI=0x4bba758f006ef09402ef31724203f316ab74e4a0
# ASRI=0x7b465370bb7a333f99edd19599eb7fb1c2d3f8d2
# SUI=0x4da82a327773965b8d4d85fa3db8249b387458e7
# PVI=0x37e15e4d6dffa9e5e320ee1ec036922e563cb76c

# Optional env vars
if [ -z "${CHAIN_ID:-}" ]; then
    # Fallback to Holesky
    CHAIN_ID=17000
fi

verify() {
    if [ "${BLOCKSCOUT_API_KEY:-}" ]; then
        forge verify-contract $1 $2 \
            --chain-id $CHAIN_ID \
            --etherscan-api-key=$BLOCKSCOUT_API_KEY \
            --verifier-url=https://eth-holesky.blockscout.com/api/ \
            --watch
    fi
    if [ "${ETHERSCAN_API_KEY:-}" ]; then
        echo "Etherscan not yet supported! Missing constructor verification";
        # forge verify-contract $1 $2 \
        #     --chain-id $CHAIN_ID \
        #     --etherscan-api-key=$ETHERSCAN_API_KEY \
        #     --watch
    fi
    if [ "${TENDERLY_URL:-}" && "${TENDERLY_API_KEY:-}" ]; then
        forge verify-contract $1 $2 \
            --chain-id $CHAIN_ID \
            --verifier-url=$TENDERLY_URL \
            --etherscan-api-key=$TENDERLY_API_KEY \
            --watch
    fi
}

verify $OPCM OPContractsManager
verify $DWI DelayedWETH
verify $OPI OptimismPortal2
verify $SCI SystemConfig
verify $LCDMI L1CrossDomainMessenger
verify $LEBI L1ERC721Bridge
verify $LSBI L1StandardBridge
verify $OMEFI OptimismMintableERC20Factory
verify $DGFI DisputeGameFactory
verify $ASRI AnchorStateRegistry
verify $SUI SuperchainConfig
verify $PVI ProtocolVersions
