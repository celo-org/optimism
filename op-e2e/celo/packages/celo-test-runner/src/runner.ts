import { Context } from "./context.ts";
import type { Config, TestCases, TestDefinitionsPerFile } from "./types.ts";
import {
  getTests,
  implementsTestCaseAsync,
  implementsTestCaseSync,
} from "./testload.ts";
import { setup, teardown } from "./setup.ts";

export async function run(config: Config) {
  const tests = await getTests(config.TestDirPath);
  Deno.test({
    name: "celo-test runner",
    permissions: { read: true, net: ["localhost", "127.0.0.1"], run: true },
    sanitizeOps: false,
    sanitizeResources: false,
    sanitizeExit: false,
    fn: async (t) => {
      await runAllTests(t, config, tests);
    },
  });
}

function filterConcurrentTests(testCases: TestCases): TestCases {
  return testCases.filter(
    (testCase) =>
      implementsTestCaseAsync(testCase.Func) &&
      testCase.ExecuteConcurrent === true,
  );
}
function filterSerialAsyncTests(testCases: TestCases): TestCases {
  return testCases.filter(
    (testCase) =>
      implementsTestCaseAsync(testCase.Func) &&
      testCase.ExecuteConcurrent === false,
  );
}
function filterSyncTests(testCases: TestCases): TestCases {
  return testCases.filter(
    (testCase) =>
      implementsTestCaseSync(testCase.Func) &&
      testCase.ExecuteConcurrent === false,
  );
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
      numChildContexts: 0,
    };
  }
  const step = {
    name: fileName,
    sanitizeOps: false,
    sanitizeResources: false,
    sanitizeExit: false,
    fn: async (t: Deno.TestContext) => {
      for (const [_, test] of asyncTests.entries()) {
        const ctx = parentCtx.child(false);
        await t.step({
          name: `${test.Name} (serial async)`,
          sanitizeOps: false,
          sanitizeResources: false,
          sanitizeExit: false,
          fn: async (t) => {
            await test.Func(t, ctx);
          },
        });
      }
      for (const [_, test] of syncTests.entries()) {
        const ctx = parentCtx.child(false);
        await t.step({
          name: `${test.Name} (serial sync)`,
          sanitizeOps: false,
          sanitizeResources: false,
          sanitizeExit: false,
          fn: (t) => {
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
    numChildContexts: filterSyncTests.length + filterSerialAsyncTests.length,
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
      numChildContexts: 0,
    };
  }
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
            name: `${test.Name} (concurrent)`,
            sanitizeOps: false,
            sanitizeResources: false,
            sanitizeExit: false,
            fn: async (t) => {
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
    numChildContexts: filterConcurrentTests.length,
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
        totalConcurrentChildContexts += concurrentDefs.numChildContexts;
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
      // we only need one account for the serial defs
      totalConcurrentChildContexts++;
    }
    const success = await t.step({
      name: `distribute funds to ${totalConcurrentChildContexts} test-acccounts`,
      fn: async (_) => {
        if (rootCtx === undefined) {
          throw Error("context is undefined");
        }
        rootCtx.clientManager.reset(totalConcurrentChildContexts);
        await rootCtx.clientManager.fundAccountsFrom(
          rootCtx.config.FunderPrivateKey,
        );
      },
    });
    if (!success) {
      // don't continue to run the tests
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
    await t.step({
      name: "teardown",
      sanitizeOps: false,
      sanitizeResources: false,
      sanitizeExit: false,
      fn: async (t1) => {
        await teardown(t1, config);
      },
    });
  }
}
