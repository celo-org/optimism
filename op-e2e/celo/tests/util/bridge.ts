import { join } from "jsr:@std/path";
import { Context } from "@celo-test/runner";
import { Address, parseEventLogs } from "viem";
import { BridgedERC20TokenPair, getContractAddress } from "@celo-test/viem";
import { expect } from "jsr:@std/expect";

export async function setupERC20BridgeToken(
  ctx: Context,
  nativeChain: any,
  bridgingAmount: BigInt,
): Promise<BridgedERC20TokenPair> {
  let pubCNative: any;
  let wlltCNative: any;
  let pubCRemote: any;
  let wlltCRemote: any;
  let mntblFactoryRemote: Address | undefined;

  if (nativeChain.id === ctx.public().l1.chain.id) {
    pubCNative = ctx.public().l1;
    wlltCNative = ctx.wallet().l1;
    pubCRemote = ctx.public().l2;
    wlltCRemote = ctx.wallet().l2;
    mntblFactoryRemote = getContractAddress(
      ctx.public().l2.chain,
      ctx.public().l2.chain, // get the factory on the remote chain
      "optimismMintableERC20Factory",
    );
    if (mntblFactoryRemote === undefined) {
      throw Error("optimismMintableERC20Factory address not known on l2");
    }
  } else if (nativeChain.id === ctx.public().l2.chain.id) {
    pubCNative = ctx.public().l2;
    wlltCNative = ctx.wallet().l2;
    pubCRemote = ctx.public().l1;
    wlltCRemote = ctx.wallet().l1;
    mntblFactoryRemote = getContractAddress(
      ctx.public().l2.chain,
      ctx.public().l1.chain, // get the factory on the remote chain
      "optimismMintableERC20Factory",
    );
    if (mntblFactoryRemote === undefined) {
      throw Error("optimismMintableERC20Factory address not known on l1");
    }
  } else {
    throw Error("native chain does not match any client chain ids");
  }

  const mintableERC20Factory = JSON.parse(
    Deno.readTextFileSync(
      join(
        ctx.config.TestDirPath,
        "../contracts/OptimismMintableERC20Factory.json",
      ),
    ),
  );
  const testBridgeToken = JSON.parse(
    Deno.readTextFileSync(
      join(ctx.config.TestDirPath, "../contracts/TestBridgeToken.json"),
    ),
  );

  let txHash = await wlltCNative.deployContract({
    abi: testBridgeToken.abi,
    args: [],
    bytecode: testBridgeToken.bytecode["object"],
  });
  console.log("deploy hash", txHash);
  const tokenDeployReceipt = await pubCNative.waitForTransactionReceipt({
    hash: txHash,
  });
  if (!tokenDeployReceipt.contractAddress) {
    throw Error("receipt didn't have contract address");
  }
  // TODO: only mint if we don't have balance?
  const nativeTokenAddress: Address = tokenDeployReceipt.contractAddress;
  // TODO: use the CREATE2 deployer,
  // so that we can reuse the token on mainnet
  // TODO: mint as often as long as we don't have the bridgingAmount balance
  txHash = await wlltCNative.writeContract({
    address: nativeTokenAddress,
    abi: testBridgeToken.abi,
    functionName: "mint100",
    args: [],
  });
  let receipt = await pubCNative.waitForTransactionReceipt({
    hash: txHash,
  });
  const { request } = await pubCRemote.simulateContract({
    account: wlltCRemote.account,
    address: mntblFactoryRemote,
    abi: mintableERC20Factory.abi,
    functionName: "createOptimismMintableERC20",
    args: [nativeTokenAddress, "TestBridgeToken", "TBT"],
  });
  console.log("tx write request", request);

  txHash = await wlltCRemote.writeContract(request);
  receipt = await pubCRemote.waitForTransactionReceipt({
    hash: txHash,
  });
  expect(receipt.status).toBe("success");
  const topics = parseEventLogs({
    abi: mintableERC20Factory.abi,
    logs: receipt.logs,
  });
  expect(topics.length).toBe(2);
  // FIXME: type
  const bridgeTokenAddress = topics[1].args!.localToken;

  // FIXME: not strictly equal, because the actual address is checksummed
  // expect(topics[1].args!.remoteToken).toBe(bridgingTokenAddressL1);
  //
  // TODO: log the chain-id with the token address
  console.log("created native token on chain", nativeTokenAddress);
  console.log("created remote token on chain", bridgeTokenAddress);

  const bridgedTokenERC20 = await pubCRemote.getERC20({
    erc20: {
      address: bridgeTokenAddress,
      chainID: pubCRemote.chain!.id,
    },
  });
  const nativeTokenERC20 = await pubCNative.getERC20({
    erc20: {
      address: nativeTokenAddress,
      chainID: pubCNative.chain!.id,
    },
  });

  return {
    bridgedToken: bridgedTokenERC20,
    nativeToken: nativeTokenERC20,
    nativeOnL1: pubCNative.chain!.id === ctx.public().l1!.chain!.id,
  };
}
