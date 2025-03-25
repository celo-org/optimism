import type { Context } from "./context.ts";
import { bindLoggerToTest } from "./logger.ts";
import type {
  Config,
  TestCase,
  TestCases,
  TestDefinitionsPerFile,
} from "./types.ts";
import {
  implementsTestCaseAsync,
  implementsTestCaseSync,
  importTestsForDirectory,
} from "./testimport.ts";
import { setup, teardown } from "./setup.ts";

export async function run(config: Config) {
  const tests = await importTestsForDirectory(config.TestDirPath);
  Deno.test({
    name: "celo-test runner",
    sanitizeOps: false,
    sanitizeResources: false,
    sanitizeExit: false,
    fn: async (t) => {
      await runAllTests(t, config, tests);
    },
  });
}

function countSkippedTests(testCases: TestCases, config: Config): number {
  return testCases.reduce(
    (count, item) => count + (skipTest(item, config) ? 1 : 0),
    0,
  );
}
function filterConcurrentTests(testCases: TestCases): TestCases {
  return testCases.filter(
    (testCase) =>
      implementsTestCaseAsync(testCase.Func) &&
      testCase.Metadata.Concurrent === true,
  );
}
function filterSerialAsyncTests(testCases: TestCases): TestCases {
  return testCases.filter(
    (testCase) =>
      implementsTestCaseAsync(testCase.Func) &&
      testCase.Metadata.Concurrent === false,
  );
}
function filterSyncTests(testCases: TestCases): TestCases {
  return testCases.filter(
    (testCase) =>
      implementsTestCaseSync(testCase.Func) &&
      testCase.Metadata.Concurrent === false,
  );
}
function skipTest(test: TestCase, config: Config): boolean {
  if (test.Metadata.OnlyRunOnL2ChainIDs === undefined) {
    return false;
  }
  const allowedIDs = new Set(test.Metadata.OnlyRunOnL2ChainIDs);
  return !allowedIDs.has(config.L2.ChainID);
}

function serialTestDefinitionsForFile(
  parentCtx: Context,
  fileName: string,
  testCases: TestCases,
): TestDefinitionsPerFile {
  const asyncTests = filterSerialAsyncTests(testCases);
  const syncTests = filterSyncTests(testCases);
  if (asyncTests.length == 0 && syncTests.length == 0) {
    return {
      name: fileName,
      step: null,
      concurrent: false,
      numActiveTestSteps: 0,
    };
  }
  let numSkippedTests = countSkippedTests(syncTests, parentCtx.config);
  numSkippedTests += countSkippedTests(asyncTests, parentCtx.config);
  const step = {
    name: fileName,
    sanitizeOps: false,
    sanitizeResources: false,
    sanitizeExit: false,
    fn: async (t: Deno.TestContext) => {
      for (const [_, test] of asyncTests.entries()) {
        const ctx = parentCtx.child(false);
        await t.step({
          name: `${test.Metadata.Name} (execution=serial)`,
          sanitizeOps: false,
          ignore: skipTest(test, ctx.config),
          sanitizeResources: false,
          sanitizeExit: false,
          fn: async (t) => {
            bindLoggerToTest(t, ctx);
            await test.Func(t, ctx);
          },
        });
      }
      for (const [_, test] of syncTests.entries()) {
        const ctx = parentCtx.child(false);
        await t.step({
          name: `${test.Metadata.Name} (execution=serial)`,
          sanitizeOps: false,
          sanitizeResources: false,
          ignore: skipTest(test, ctx.config),
          sanitizeExit: false,
          fn: (t) => {
            bindLoggerToTest(t, ctx);
            test.Func(t, ctx);
          },
        });
      }
    },
  };
  return {
    name: fileName,
    step: step,
    concurrent: false,
    numActiveTestSteps: syncTests.length + asyncTests.length - numSkippedTests,
  };
}

function concurrentTestDefinitionsForFile(
  parentCtx: Context,
  fileName: string,
  testCases: TestCases,
): TestDefinitionsPerFile {
  const tests = filterConcurrentTests(testCases);
  if (tests.length == 0) {
    return {
      name: fileName,
      step: null,
      concurrent: true,
      numActiveTestSteps: 0,
    };
  }
  const numSkippedTests = countSkippedTests(tests, parentCtx.config);
  const step = {
    name: fileName,
    sanitizeOps: false,
    sanitizeResources: false,
    sanitizeExit: false,
    fn: async (t: Deno.TestContext) => {
      const testPromises: Array<Promise<boolean>> = [];
      for (const [_, test] of tests.entries()) {
        const ctx = parentCtx.child(true);
        testPromises.push(
          t.step({
            name: `${test.Metadata.Name} (execution=concurrent)`,
            sanitizeOps: false,
            ignore: skipTest(test, ctx.config),
            sanitizeResources: false,
            sanitizeExit: false,
            fn: async (t) => {
              bindLoggerToTest(t, ctx);
              await test.Func(t, ctx);
            },
          }),
        );
      }
      await Promise.allSettled(testPromises);
    },
  };
  return {
    name: fileName,
    step: step,
    concurrent: true,
    numActiveTestSteps: tests.length - numSkippedTests,
  };
}
async function runAllTests(
  t: Deno.TestContext,
  config: Config,
  tests: Record<string, TestCases>,
) {
  let rootCtx: Context | undefined = undefined;
  try {
    const setupSuccess = await t.step({
      name: "setup",
      sanitizeOps: false,
      sanitizeResources: false,
      sanitizeExit: false,
      fn: async (t) => {
        rootCtx = await setup(t, config);
      },
    });
    if (!setupSuccess) {
      throw Error("setup failed");
    }
    if (rootCtx === undefined) {
      throw Error("parent test context couldn't be set up");
    }
    const concurrentTests: Array<Deno.TestStepDefinition> = [];
    const serialTests: Array<Deno.TestStepDefinition> = [];
    let totalConcurrentChildContexts = 0;
    for (const [fileName, testCases] of Object.entries(tests)) {
      const concurrentDefs = concurrentTestDefinitionsForFile(
        rootCtx,
        fileName,
        testCases,
      );
      if (concurrentDefs.step !== null) {
        concurrentTests.push(concurrentDefs.step);
        // we need a new account per each concurrently executed test
        // to avoid nonce overlap
        totalConcurrentChildContexts += concurrentDefs.numActiveTestSteps;
      }
      const serialDefs = serialTestDefinitionsForFile(
        rootCtx,
        fileName,
        testCases,
      );
      if (serialDefs.step !== null) {
        serialTests.push(serialDefs.step);
      }
    }
    if (serialTests.length > 0) {
      // we only need one account for the serially executed tests
      totalConcurrentChildContexts++;
    }
    const success = await t.step({
      name: `distribute funds to ${totalConcurrentChildContexts} test-acccounts`,
      fn: async (_) => {
        if (rootCtx === undefined) {
          throw Error("context is undefined");
        }
        rootCtx.resetClients(totalConcurrentChildContexts);
        await rootCtx.clientManager.fundAccountsFrom(
          rootCtx.config.FunderPrivateKey,
        );
      },
    });
    if (!success) {
      throw Error("couldn't distribute test acccount funds");
    }

    const steps: Array<Promise<boolean>> = [];
    if (concurrentTests.length > 0) {
      steps.push(
        t.step({
          name: "concurrently running tests",
          fn: async (_) => {
            await Promise.allSettled(
              concurrentTests.map((test) => t.step(test)),
            );
          },
        }),
      );
    }
    if (serialTests.length > 0) {
      steps.push(
        t.step({
          name: "serially running tests",
          fn: async (_) => {
            for (const test of serialTests) {
              await t.step(test);
            }
          },
        }),
      );
    }
    await Promise.allSettled(steps);
  } finally {
    // TODO: refund the funder account from all contexts
    await t.step({
      name: "teardown",
      sanitizeOps: false,
      sanitizeResources: false,
      sanitizeExit: false,
      fn: async (t1) => {
        if (rootCtx !== undefined) {
          console.log("writing artifacts file");
          await rootCtx.logger.flush();
        }
        await teardown(t1, config);
      },
    });
  }
}
