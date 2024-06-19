export const withdraw = async function(args, config) {
  const initiateHash = await config.client.l2.wallet.initiateWithdrawal({
    request: {
      gas: args.gas,
      to: args.to,
      value: args.amount,
    },
  })

  console.log('withdraw hash', initiateHash)

  const receipt = await config.client.l2.public.waitForTransactionReceipt({
    hash: initiateHash,
  })
  console.log('receipt', receipt)

  // NOTE: this function requires the mulitcall contract to be deployed
  // on the L1 chain.
  const { output, withdrawal } = await config.client.l1.public.waitToProve({
    receipt,
    targetChain: config.client.l2.public.chain,
  })
  console.log('output', output)
  console.log('withdrawal', withdrawal)

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
    console.log('prove receipt not successful')
    return false
  }

  await config.client.l1.public.waitToFinalize({
    withdrawalHash: withdrawal.withdrawalHash,
    targetChain: config.client.l2.public.chain,
  })

  // the waitToFinalize has been inaccurate,
  // causing a revert with:
  // "OtimismPortal: proven withdrawal finalization period has not elapsed"
  await new Promise((r) => setTimeout(r, 2 * 1_000))

  const finalizeHash = await config.client.l1.wallet.finalizeWithdrawal({
    targetChain: config.client.l2.public.chain,
    withdrawal,
  })

  const finalizeReceipt =
    await config.client.l1.public.waitForTransactionReceipt({
      hash: finalizeHash,
    })
  console.log(finalizeReceipt)

  // TODO: finalize
  return finalizeReceipt.status == 'success'
}
