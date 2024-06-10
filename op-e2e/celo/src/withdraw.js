export const withdraw = async function (args, config) {
  const initiateHash = await config.client.l2.wallet.initiateWithdrawal({
    request: {
      gas: args.gas,
      to: args.to,
      value: args.amount,
    },
  })
  const receipt = await config.client.l2.public.waitForTransactionReceipt({
    hash: initiateHash,
  })

  const l2GasPayment =
    receipt.gasUsed * receipt.effectiveGasPrice + receipt.l1Fee

  // FIXME: this blocks longer, the longer the devnet is running, see
  // https://github.com/ethereum-optimism/optimism/issues/7668
  // NOTE: this function requires the mulitcall contract to be deployed
  // on the L1 chain.
  const { output, withdrawal } = await config.client.l1.public.waitToProve({
    receipt,
    targetChain: config.client.l2.public.chain,
  })
  //

  const proveWithdrawalArgs =
    await config.client.l2.public.buildProveWithdrawal({
      output,
      withdrawal,
    })
  const proveHash =
    await config.client.l1.wallet.proveWithdrawal(proveWithdrawalArgs)

  const proveReceipt = await config.client.l1.public.waitForTransactionReceipt({
    hash: proveHash,
  })
  if (proveReceipt.status != 'success') {
    return {
      success: false,
      l2GasPayment: l2GasPayment,
    }
  }

  await config.client.l1.public.waitToFinalize({
    withdrawalHash: withdrawal.withdrawalHash,
    targetChain: config.client.l2.public.chain,
  })

  const finalizeHash = await config.client.l1.wallet.finalizeWithdrawal({
    targetChain: config.client.l2.public.chain,
    withdrawal,
  })

  const finalizeReceipt =
    await config.client.l1.public.waitForTransactionReceipt({
      hash: finalizeHash,
    })

  return {
    success: finalizeReceipt.status == 'success',
    l2GasPayment: l2GasPayment,
  }
}
