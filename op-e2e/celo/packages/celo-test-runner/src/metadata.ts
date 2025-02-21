import "npm:reflect-metadata";
import type { TestFunc } from "./types.ts";

export type TestMetadata = {
  Name: string;
  Concurrent: boolean;
  OnlyRunOnL2ChainIDs: Array<number> | undefined;
};

const testMetadataKey = "test-metadata";

// when this decorator is not used to wrap a test-function,
// the default values for the metadata will be used (see getTestOptions).
export function addTestOptions(metadata: TestMetadata) {
  return <T extends TestFunc>(fn: T): T => {
    Reflect.defineMetadata(testMetadataKey, metadata, fn);
    return fn;
  };
}
export function getTestOptions<T extends TestFunc>(fn: T): TestMetadata {
  let meta = Reflect.getMetadata(testMetadataKey, fn);
  if (meta === undefined) {
    // default values when the function is not decorated with the
    // metadata adder:
    meta = {
      RunConcurrent: false,
      Name: fn.name,
      OnlyRunOnL2ChainID: undefined,
    };
  }
  return meta;
}
