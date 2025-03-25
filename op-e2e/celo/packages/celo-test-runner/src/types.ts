import type { Context } from "./context.ts";
import type { TestMetadata } from "./metadata.ts";
import type { Hex } from "viem";
import "jsr:@std/dotenv/load";

export type ChainConfig = {
  RPCURL: URL;
  ChainID: number;
};

export type Config = {
  L1: ChainConfig;
  L2: ChainConfig;
  FunderPrivateKey: Hex;
  UseFaultproofSystem: boolean;
  UseAltDA: boolean;
  SpawnDevnet: boolean;

  TestDirPath: string;
  ArtifactsDirPath: string;
  MonorepoPath: string;
  ContractAddressesL1FilePath: string;
  ContractAddressesL2FilePath: string | undefined;
  AccountsSeedPhrase: string;
};

export type TestFunc = TestFuncAsync | TestFuncSync;
export type TestFuncAsync = (
  t: Deno.TestContext,
  ctx: Context,
) => Promise<boolean>;

export type TestFuncSync = (t: Deno.TestContext, ctx: Context) => boolean;

export type TestCase = {
  File: string;
  Metadata: TestMetadata;
  Func: TestFuncSync | TestFuncAsync;
};

export type TestCases = Array<TestCase>;

export type TestDefinitionsPerFile = {
  name: string;
  step: Deno.TestStepDefinition | null;
  concurrent: boolean;
  numActiveTestSteps: number;
};
