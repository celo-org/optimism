export const withdraw = async function(args, config) {
  const initiateHash = await config.client.l2.wallet.initiateWithdrawal({
    request: {
      gas: args.gas,
      to: args.to,
      value: args.amount,
    },
  })
  const receipt = await config.client.l2.public.waitForTransactionReceipt({
    hash: initiateHash,
    timeout: 30_000,
  })

  const l2GasPayment =
    receipt.gasUsed * receipt.effectiveGasPrice + receipt.l1Fee

  // NOTE: this function requires the mulitcall3 contract to be deployed
  // on the L1 chain.
  const { output, game, withdrawal } =
    await config.client.l1.public.waitToProve({
      receipt,
      targetChain: config.client.l2.public.chain,
    })

  const proveWithdrawalArgs =
    await config.client.l2.public.buildProveWithdrawal({
      output,
      game,
      withdrawal,
    })
  const proveHash =
    await config.client.l1.wallet.proveWithdrawal(proveWithdrawalArgs)

  const proveReceipt = await config.client.l1.public.waitForTransactionReceipt({
    hash: proveHash,
    timeout: 30_000,
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

  // HACK: the waitToFinalize does not seem to calculate the wait time
  // correctly..., lets hardcode a wait time for now to see if it can work.
  // In theory viem is not waiting an additional DISPUTE_GAME_FINALITY_DELAY_SECONDS.
  // The current default value for this is 6, but this was not enough in manual testing.
  // TODO: fix this upstream in viem...
  await new Promise((res) => setTimeout(res, 16 * 1000))
  const finalizeHash = await config.client.l1.wallet.finalizeWithdrawal({
    targetChain: config.client.l2.public.chain,
    withdrawal,
  })

  const finalizeReceipt =
    await config.client.l1.public.waitForTransactionReceipt({
      hash: finalizeHash,
      timeout: 30_000,
    })

  return {
    success: finalizeReceipt.status == 'success',
    l2GasPayment: l2GasPayment,
  }
}
