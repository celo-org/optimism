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
  const getEnvValueAndKey = (
    key: string,
    optional: boolean,
  ): [string, string] => {
    const fullKey = `${prefix}_${key}`;
    if (!(fullKey in env)) {
      if (optional === true) {
        return ["", fullKey];
      }
      throw new Error(`Environment variable ${fullKey} is missing.`);
    }
    return [env[fullKey], fullKey];
  };
  const getEnvValueString = (key: string, optional: boolean): string => {
    const [val, _] = getEnvValueAndKey(key, optional);
    return val;
  };
  const getEnvValueBool = (key: string): boolean => {
    const [val, fullKey] = getEnvValueAndKey(key, false);
    const hexVal = stringToBool(val);
    if (hexVal === undefined) {
      throw new Error(
        `${fullKey} has to be a boolean value ('true'/'false' or '0'/'1')`,
      );
    }
    return hexVal;
  };
  const getEnvValueHex = (key: string): Hex => {
    const [val, fullKey] = getEnvValueAndKey(key, false);
    if (!isHex(val)) {
      throw new Error(
        `${fullKey} has to be a hex encoded value ('0x' prefixed)`,
      );
    }
    return val;
  };

  const l2ContractAddressesPath = getEnvValueString(
    "ADDRESSES_L2_FILEPATH",
    true,
  );
  return {
    L1: {
      RPCURL: new URL(getEnvValueString("L1_RPCURL", false)),
      ChainID: parseInt(getEnvValueString("L1_CHAINID", false), 10),
    },
    L2: {
      RPCURL: new URL(getEnvValueString("L2_RPCURL", false)),
      ChainID: parseInt(getEnvValueString("L2_CHAINID", false), 10),
    },
    SpawnDevnet: getEnvValueBool("SPAWN_DEVNET"),
    UseAltDA: getEnvValueBool("USE_ALTDA"),
    UseFaultproofSystem: getEnvValueBool("USE_FAULTPROOFS"),
    FunderPrivateKey: getEnvValueHex("FUNDER_PRIVATEKEY"),
    ContractAddressesL1FilePath: resolve(
      getEnvValueString("ADDRESSES_L1_FILEPATH", false),
    ),
    ContractAddressesL2FilePath:
      l2ContractAddressesPath && resolve(l2ContractAddressesPath),
    AccountsSeedPhrase: getEnvValueString("ACCOUNTS_SEEDPHRASE", false),
    TestDirPath: resolve(getEnvValueString("TESTDIRPATH", false)),
    ArtifactsDirPath: resolve(getEnvValueString("ARTIFACTSDIRPATH", false)),
    MonorepoPath: resolve(getEnvValueString("MONOREPOPATH", false)),
  };
}

export const DefaultEnvPrefix = "CELOTEST_";
