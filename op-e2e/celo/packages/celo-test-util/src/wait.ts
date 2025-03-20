import { getGames, getL2Output } from "viem/op-stack";

const sleepMaxWaitTimeMs = 1000;

export function sleep(milliseconds: number): Promise<void> {
  return sleepUntil(Date.now() + milliseconds);
}
export function sleepSeconds(seconds: number): Promise<void> {
  return sleep(seconds * 1000);
}

export function sleepUntil(targetTime: Date | number): Promise<void> {
  const target =
    typeof targetTime === "number" ? targetTime : targetTime.getTime();

  return new Promise((resolve) => {
    const check = () => {
      const now = Date.now();
      if (now >= target) {
        resolve();
      } else {
        // calculate remaining time, but cap the delay to a maximum value
        const delay = Math.min(target - now, sleepMaxWaitTimeMs);
        // recursively call the time-checking with a timeout
        setTimeout(check, delay);
      }
    };
    check();
  });
}

export async function pollFunction<T>(
  fn: () => Promise<T>,
  until: (value: T | null, err: Error | null) => boolean | undefined,
  initialPollInterval: number,
  timeout: number | undefined,
  exponentialBackoff: boolean,
): Promise<T | null> {
  const start = Date.now();
  let err: Error | null = null;
  let result: T | null = null;
  const pollScalingFactor = exponentialBackoff ? 1.2 : 1;
  let lastPollTick = start;
  let currentPollInterval = initialPollInterval;
  const maxPollInterval = initialPollInterval * 100;

  if (typeof until !== "function") {
    throw Error("passed in `until` parameter is not a function");
  }
  const waitUntilNextTick = async function (): Promise<void> {
    const timeNow = Date.now();
    const timePassed = timeNow - lastPollTick;

    let pollTime = currentPollInterval - timePassed;
    if (pollTime < 0) {
      // we should have polled already, don't wait
      return;
    }
    if (timeout !== undefined) {
      const timeUntilTimeout = start + timeout - timeNow;
      if (timeUntilTimeout < 0) {
        return;
      }
      if (timeUntilTimeout < pollTime) {
        pollTime = timeUntilTimeout;
      }
    }

    await sleep(pollTime);
    lastPollTick = Date.now();
    currentPollInterval = currentPollInterval * pollScalingFactor;
    if (currentPollInterval > maxPollInterval) {
      currentPollInterval = maxPollInterval;
    }
  };

  while (timeout === undefined || Date.now() - start < timeout) {
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
    //NOTE: the function will never
    // throw an error before the timeout, or never throw an error at all
    // when no timeout is given.
    // That means that the 'until' function
    // has to throw or return true if the poll-loop
    // should get canceled
    if (until(result, err) === true) {
      return result;
    }
    await waitUntilNextTick();
  }
  if (err) {
    throw new Error(`timeout reached polling function: ${err.message}`, {
      cause: err,
    });
  }
  throw new Error(`timeout reached polling function`);
}

export async function waitClientReachable(
  client: { getChainId: () => Promise<number> },
  timeoutSeconds: number,
): Promise<Error | null> {
  const until = (val: number | null, _err: Error | null): boolean => {
    return typeof val === "number";
  };
  await pollFunction(
    client.getChainId,
    until,
    500,
    timeoutSeconds * 1000,
    false,
  );
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
    false,
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
  await pollFunction(
    client.getBlockNumber,
    until,
    500,
    timeoutSeconds * 1000,
    false,
  );
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
    true,
  );
  return null;
}

export async function waitUntilTwoGames(publicClients: any, timeout: number) {
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
      games !== null ? games.length >= 2 : false,
    500,
    timeout * 1000,
    true,
  );
  return null;
}
