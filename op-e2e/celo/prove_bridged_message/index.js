#!/usr/bin/env node

/* Receive an L2 txhash as CLI parameter and prove+finalize it on L1 */

const sdk = require('@eth-optimism/sdk')
const { EXDEV } = require('constants')
const ethers = require('ethers')
const fs = require('fs')

const txhash = process.argv[2]
console.log('txhash:', txhash)
const raw_contract_addresses = JSON.parse(fs.readFileSync('../../.devnet/addresses.json'))

const contract_addresses = {
  AddressManager: raw_contract_addresses.AddressManager,
  L1CrossDomainMessenger: raw_contract_addresses.L1CrossDomainMessengerProxy,
  L1StandardBridge: raw_contract_addresses.L1StandardBridgeProxy,
  OptimismPortal: raw_contract_addresses.OptimismPortalProxy,
  L2OutputOracle: raw_contract_addresses.L2OutputOracleProxy,

  // These are only needed for backwards compatibility and are not used
  StateCommitmentChain: "0x".padEnd(42, "0"),
  CanonicalTransactionChain: "0x".padEnd(42, "0"),
  BondManager: "0x".padEnd(42, "0"),
}

const waitForStatus = async function(expectedStatus) {
  let status;
  while (status !== expectedStatus) {
    status = await crossChainMessenger.getMessageStatus(txhash)
    if (status === expectedStatus) {
      return
    }
    console.log("Waiting for message status", expectedStatus, ", current status is", status)
    await new Promise(r => setTimeout(r, 2000))
  }
}

const main = async function () {
  // crossChainMessenger configuration
  l2Provider = new ethers.providers.JsonRpcProvider(process.env.ETH_RPC_URL)
  l1Provider = new ethers.providers.JsonRpcProvider(process.env.ETH_RPC_URL_L1)
  l1Signer = new ethers.Wallet(process.env.ACC_PROVER_PRIVKEY, l1Provider)
  const l1ChainId = (await l1Provider.getNetwork()).chainId
  const l2ChainId = (await l2Provider.getNetwork()).chainId
  crossChainMessenger = new sdk.CrossChainMessenger({
    l1ChainId: l1ChainId,
    l2ChainId: l2ChainId,
    l1SignerOrProvider: l1Signer,
    l2SignerOrProvider: l2Provider,
    contracts: {
      l1: contract_addresses
    }
  })

  // Prove
  await waitForStatus(sdk.MessageStatus.READY_TO_PROVE)
  const proveTx = await crossChainMessenger.proveMessage(txhash)
  console.log('prove message tx:', proveTx.hash)

  // Finalize
  await waitForStatus(sdk.MessageStatus.READY_FOR_RELAY)
  const finalizeTx = await crossChainMessenger.finalizeMessage(txhash)
  console.log('finalize message tx:', finalizeTx.hash)

  // Wait for tx confirmation before exit
  await finalizeTx.wait()
  console.log('Withdraw proven for tx', txhash)
}

main()
