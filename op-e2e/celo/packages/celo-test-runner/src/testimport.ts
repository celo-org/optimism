import type {
  TestCase,
  TestCases,
  TestFuncAsync,
  TestFuncSync,
} from "./types.ts";
import { getTestOptions } from "./metadata.ts";

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

export async function importTestsForDirectory(
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
      Object.entries(module).forEach(([_, value]) => {
        if (implementsTestCaseAsync(value) || implementsTestCaseSync(value)) {
          const metadata = getTestOptions(value);
          const testCase: TestCase = {
            File: entry.name,
            Metadata: metadata,
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
