import type { Context } from "./context.ts";
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
  MonorepoPath: string;
};

export type TestFuncAsync = (
  t: Deno.TestContext,
  ctx: Context,
) => Promise<boolean>;

export type TestFuncSync = (t: Deno.TestContext, ctx: Context) => boolean;

export type TestCase = {
  File: string;
  Name: string;
  ExecuteConcurrent: boolean;
  Func: TestFuncSync | TestFuncAsync;
};

export type TestCases = Array<TestCase>;

export type TestDefinitionsPerFile = {
  name: string;
  step: Deno.TestStepDefinition | null;
  concurrent: boolean;
  numChildContexts: number;
};
