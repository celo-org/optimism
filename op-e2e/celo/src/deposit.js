import { getL2TransactionHashes } from 'viem/op-stack'
import { OptimismPortalABI } from './OptimismPortal.js'

// public client functionality
export async function constructDepositCustomGas(client, parameters) {
  const {
    account,
    chain = client.chain,
    gas,
    maxFeePerGas,
    maxPriorityFeePerGas,
    nonce,
    request: {
      data = '0x',
      gas: l2Gas,
      isCreation = false,
      mint,
      to = '0x',
      value,
    },
    targetChain,
  } = parameters

  const portalAddress = (() => {
    if (parameters.portalAddress) return parameters.portalAddress
    if (chain) return targetChain.contracts.portal[chain.id].address
    return Object.values(targetChain.contracts.portal)[0].address
  })()
  const callArgs = {
    account: account,
    abi: OptimismPortalABI,
    address: portalAddress,
    chain,
    functionName: 'depositERC20Transaction',
    /// @notice Entrypoint to depositing an ERC20 token as a custom gas token.
    ///         This function depends on a well formed ERC20 token. There are only
    ///         so many checks that can be done on chain for this so it is assumed
    ///         that chain operators will deploy chains with well formed ERC20 tokens.
    /// @param _to         Target address on L2.
    /// @param _mint       Units of ERC20 token to deposit into L2.
    /// @param _value      Units of ERC20 token to send on L2 to the recipient.
    /// @param _gasLimit   Amount of L2 gas to purchase by burning gas on L1.
    /// @param _isCreation Whether or not the transaction is a contract creation.
    /// @param _data       Data to trigger the recipient with.
    args: [
      isCreation ? zeroAddress : to,
      mint ?? value ?? 0n,
      value ?? mint ?? 0n,
      l2Gas,
      isCreation,
      data,
    ],
    maxFeePerGas,
    maxPriorityFeePerGas,
    nonce,
  }
  const gas_ =
    typeof gas !== 'number' && gas !== null
      ? await client.estimateContractGas(callArgs)
      : undefined
  callArgs.gas = gas_
  const result = client.simulateContract(callArgs)
  return { result: result, args: callArgs }
}

export async function deposit(args, config) {
  var spentGas = BigInt(0)
  const depositArgs = await config.client.l2.public.buildDepositTransaction({
    mint: args.mint,
    to: args.to,
  })

  const celoToken = await config.client.l1.public.getERC20({
    erc20: {
      address: config.addresses.CustomGasTokenProxy,
      chainID: config.client.l1.public.chain.id,
    },
  })
  const portalAddress =
    config.client.l2.public.chain.contracts.portal[
      config.client.l1.public.chain.id
    ].address
  const approve = await config.client.l1.wallet.simulateERC20Approve({
    amount: { amount: args.mint, token: celoToken },
    spender: portalAddress,
  })
  if (!approve.result) {
    return {
      success: false,
      l1GasPayment: spentGas,
    }
  }

  const approveHash = await config.client.l1.wallet.writeContract(
    approve.request
  )
  // Wait for the L1 transaction to be processed.
  const approveReceipt =
    await config.client.l1.public.waitForTransactionReceipt({
      hash: approveHash,
    })

  spentGas += approveReceipt.gasUsed * approveReceipt.effectiveGasPrice
  const dep =
    await config.client.l1.public.prepareDepositGasPayingTokenERC20(depositArgs)
  const hash = await config.client.l1.wallet.writeContract(dep.args)

  // Wait for the L1 transaction to be processed.
  const receipt = await config.client.l1.public.waitForTransactionReceipt({
    hash: hash,
  })

  spentGas += receipt.gasUsed * receipt.effectiveGasPrice

  // Get the L2 transaction hash from the L1 transaction receipt.
  const [l2Hash] = getL2TransactionHashes(receipt)

  // Wait for the L2 transaction to be processed.
  const l2Receipt = await config.client.l2.public.waitForTransactionReceipt({
    hash: l2Hash,
  })

  return {
    success: l2Receipt.status == 'success',
    l1GasPayment: spentGas,
  }
}
