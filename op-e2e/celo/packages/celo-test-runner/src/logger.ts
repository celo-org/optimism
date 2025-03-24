// deno-lint-ignore-file no-explicit-any
import { ensureDir } from "jsr:@std/fs/ensure-dir";
import { join } from "jsr:@std/path";
import type { Context } from "./context.ts";
import type { Address } from "viem";

interface TestInfo {
  l1Info: {
    address: Address;
    chainId: number;
    rpcUrl: string;
  };
  l2Info: {
    address: Address;
    chainId: number;
    rpcUrl: string;
  };
  concurrent: boolean;
}

interface LogEntry {
  timestamp: string;
  message: string;
  data: any;
}

export function stringifyWithBigInt(obj: any): string {
  return JSON.stringify(
    obj,
    (_k, v) => (typeof v === "bigint" ? v.toString() : v),
    2,
  );
}

export function parseWithBigInt(json: string): any {
  return JSON.parse(json, (_k, v) =>
    typeof v === "string" && /^\d{15,}$/.test(v) ? BigInt(v) : v,
  );
}

export function bindLoggerToTest(t: Deno.TestContext, ctx: Context) {
  const logger = ctx.logger;
  logger.addMetadata(t.name, ctx);
  ctx.injectArtifactStore(t.name, logger.store.bind(logger));
}

export class TestLogger {
  private rootDir: string;
  private logsByTest: Record<string, LogEntry[]> = {};
  private testInfoByTest: Record<string, TestInfo> = {};

  constructor(rootDir: string) {
    this.rootDir = rootDir;
  }

  addMetadata(id: string, ctx: Context) {
    const info = {
      l1Info: {
        address: ctx.wallet().l1.account!.address,
        chainId: ctx.config.L1.ChainID,
        rpcUrl: ctx.config.L1.RPCURL.toString(),
      },
      l2Info: {
        address: ctx.wallet().l2.account!.address,
        chainId: ctx.config.L2.ChainID,
        rpcUrl: ctx.config.L2.RPCURL.toString(),
      },
      concurrent: ctx.concurrent,
    };
    this.testInfoByTest[id] = info;
  }

  store(id: string, message: string, data: any) {
    if (!this.logsByTest[id]) {
      this.logsByTest[id] = [];
    }
    this.logsByTest[id].push({
      timestamp: new Date().toISOString(),
      message,
      data,
    });
  }

  async flush() {
    const tests: Record<string, { info?: TestInfo; logs: LogEntry[] }> = {};

    for (const [testName, logs] of Object.entries(this.logsByTest)) {
      tests[testName] = {
        info: this.testInfoByTest[testName],
        logs,
      };
    }

    const artifact = {
      tests,
    };

    await ensureDir(this.rootDir);

    const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
    const outFile = join(this.rootDir, `test_logs_${timestamp}.json`);

    await Deno.writeTextFile(outFile, stringifyWithBigInt(artifact));

    console.log(`all test logs flushed to ${outFile}`);
  }
}
