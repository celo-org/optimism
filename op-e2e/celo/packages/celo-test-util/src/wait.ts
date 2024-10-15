import { getGames, getL2Output } from "viem/op-stack";

export async function sleep(time: number): Promise<void> {
  await new Promise((r) => setTimeout(r, time));
}

export async function pollFunction<T>(
  fn: () => Promise<T>,
  until: (value: T | null, err: Error | null) => boolean | undefined,
  pollInterval: number,
  timeout: number,
): Promise<T | null> {
  const start = Date.now();
  let err: Error | null = null;
  let result: T | null = null;

  while (Date.now() - start < timeout) {
    err = null;
    result = null;
    try {
      result = await fn();
    } catch (error) {
      if (error instanceof Error) {
        err = error;
      } else {
        console.log("caught unknown error type", err);
      }
    }
    //FIXME: like this the function will never
    // throw an error before the timeout... is that desired?
    if (typeof until === "function") {
      if (until(result, err) === true) {
        return result;
      }
    }
    await sleep(pollInterval);
  }
  if (err) {
    throw err;
  }
  return null;
}

export async function waitClientReachable(
  client: { getChainId: () => Promise<number> },
  timeoutSeconds: number,
): Promise<Error | null> {
  const until = (val: number | null, _err: Error | null): boolean => {
    return typeof val === "number";
  };
  await pollFunction(client.getChainId, until, 500, timeoutSeconds * 1000);
  return null;
}

export async function waitClientNotSyncing(
  // deno-lint-ignore no-explicit-any
  client: any,
  timeoutSeconds: number,
): Promise<Error | null> {
  await pollFunction(
    async (): Promise<boolean> => {
      return await client.request({ method: "eth_syncing" });
    },
    (val: boolean | null, _err: Error | null): boolean => val === false,
    500,
    timeoutSeconds * 1000,
  );
  return null;
}

export async function waitClientReturnsBlockNum(
  client: { getBlockNumber: () => Promise<bigint> },
  timeoutSeconds: number,
): Promise<Error | null> {
  const until = (val: bigint | null, _err: Error | null): boolean => {
    return typeof val === "bigint";
  };
  await pollFunction(client.getBlockNumber, until, 500, timeoutSeconds * 1000);
  return null;
}

export async function waitInitialL2OracleOutput(
  publicClients: any,
  timeout: number,
) {
  const fn = async () => {
    // @ts-ignore: allow anonymous type passing until the celo-e2e package
    // is ported to TS
    const games = await getL2Output(publicClients.l1, {
      targetChain: publicClients.l2.chain,
      // set 0n here, since we just want to wait so that
      // the contract returns any games at all, otherwise
      // viem can run into an uncaught index-error during bridging
      l2BlockNumber: 0n,
    });
    return games;
  };

  await pollFunction(
    fn,
    (l2Output: any | null, _err: Error | null) => !!l2Output,
    500,
    timeout * 1000,
  );
  return null;
}

export async function waitInitialGame(publicClients: any, timeout: number) {
  const fn = async () => {
    // @ts-ignore: allow anonymous type passing until the celo-e2e package
    // is ported to TS
    return await getGames(publicClients.l1, {
      targetChain: publicClients.l2.chain,
    });
  };

  await pollFunction(
    fn,
    (games: Array<any> | null, _err: Error | null) =>
      games !== null ? games.length > 0 : false,
    500,
    timeout * 1000,
  );
  return null;
}
