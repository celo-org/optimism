import type {
  TestCase,
  TestCases,
  TestFuncAsync,
  TestFuncSync,
} from "./types.ts";

// Utility function to check if a function matches the TestCase interface.
// Note that TS is lacking runtime checks for type definitions, and
// the check here is very basic (doesn't check for type signature, only length)
function implementsTestCase(func: unknown, sync: boolean): boolean {
  if (!(typeof func === "function" && func.length === 2)) {
    return false;
  }
  if (sync) {
    return func.constructor.name === "Function";
  } else {
    return func.constructor.name === "AsyncFunction";
  }
}

export function implementsTestCaseSync(func: unknown): func is TestFuncSync {
  return implementsTestCase(func, true);
}
export function implementsTestCaseAsync(func: unknown): func is TestFuncAsync {
  return implementsTestCase(func, false);
}
function shouldExecuteConcurrent(name: string): boolean {
  return name.endsWith("Concurrent");
}

export async function getTests(
  directory: string,
): Promise<Record<string, TestCases>> {
  const testCasesPerFile: Record<string, TestCases> = {};
  // expects a flat file directory, no nested folders
  for await (const entry of Deno.readDir(directory)) {
    const testCases: TestCases = [];
    if (
      entry.isFile &&
      (entry.name.endsWith(".ts") || entry.name.endsWith(".js"))
    ) {
      const absolutePath = await Deno.realPath(`${directory}/${entry.name}`); // Resolve to an absolute path
      const module = await import(absolutePath);
      // Filter and merge functions that implement the TestCase interface
      Object.entries(module).forEach(([key, value]) => {
        if (implementsTestCaseAsync(value) || implementsTestCaseSync(value)) {
          const testCase: TestCase = {
            Name: key,
            File: entry.name,
            ExecuteConcurrent: shouldExecuteConcurrent(key),
            Func: value,
          };
          testCases.push(testCase);
        }
      });
    }
    if (testCases.length !== 0) {
      testCasesPerFile[entry.name] = testCases;
    }
  }
  return testCasesPerFile;
}
