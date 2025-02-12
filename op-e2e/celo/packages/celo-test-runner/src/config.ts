import "jsr:@std/dotenv/load";
import type { Config } from "./types.ts";
import { type Hex, isHex } from "viem";
import { resolve } from "jsr:@std/path";

function stringToBool(value: string): boolean | undefined {
  const normalized = value.trim().toLowerCase();
  if (normalized === "true" || normalized === "1") {
    return true;
  }
  if (normalized === "false" || normalized === "0") {
    return false;
  }
  return undefined;
}

export function parseConfigWithPrefixFromEnv(
  env: Record<string, string>,
  prefix: string,
): Config {
  const getEnvValueAndKey = (key: string): [string, string] => {
    const fullKey = `${prefix}_${key}`;
    if (!(fullKey in env)) {
      throw new Error(`Environment variable ${fullKey} is missing.`);
    }
    return [env[fullKey], fullKey];
  };
  const getEnvValueString = (key: string): string => {
    const [val, _] = getEnvValueAndKey(key);
    return val;
  };
  const getEnvValueBool = (key: string): boolean => {
    const [val, fullKey] = getEnvValueAndKey(key);
    const hexVal = stringToBool(val);
    if (hexVal === undefined) {
      throw new Error(
        `${fullKey} has to be a boolean value ('true'/'false' or '0'/'1')`,
      );
    }
    return hexVal;
  };
  const getEnvValueHex = (key: string): Hex => {
    const [val, fullKey] = getEnvValueAndKey(key);
    if (!isHex(val)) {
      throw new Error(
        `${fullKey} has to be a hex encoded value ('0x' prefixed)`,
      );
    }
    return val;
  };

  return {
    L1: {
      RPCURL: new URL(getEnvValueString("L1_RPCURL")),
      ChainID: parseInt(getEnvValueString("L1_CHAINID"), 10),
    },
    L2: {
      RPCURL: new URL(getEnvValueString("L2_RPCURL")),
      ChainID: parseInt(getEnvValueString("L2_CHAINID"), 10),
    },
    SpawnDevnet: getEnvValueBool("SPAWN_DEVNET"),
    UseAltDA: getEnvValueBool("USE_ALTDA"),
    UseFaultproofSystem: getEnvValueBool("USE_FAULTPROOFS"),
    FunderPrivateKey: getEnvValueHex("FUNDER_PRIVATEKEY"),
    ContractAddressesFilePath: resolve(getEnvValueString("ADDRESSES_FILEPATH")),
    AccountsSeedPhrase: resolve(getEnvValueString("ACCOUNTS_SEEDPHRASE")),
    TestDirPath: resolve(getEnvValueString("TESTDIRPATH")),
    MonorepoPath: resolve(getEnvValueString("MONOREPOPATH")),
  };
}

export const DefaultEnvPrefix = "CELOTEST_";
