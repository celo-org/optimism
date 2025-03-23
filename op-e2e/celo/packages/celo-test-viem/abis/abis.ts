import { join } from "jsr:@std/path";
const __dirname = new URL(".", import.meta.url).pathname;

export const standardBridge = JSON.parse(
  Deno.readTextFileSync(join(__dirname, "StandardBridge.json")),
);
