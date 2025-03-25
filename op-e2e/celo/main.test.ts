import { parseConfigWithPrefixFromEnv, run } from "@celo-test/runner";

await run(parseConfigWithPrefixFromEnv(Deno.env.toObject(), "CELOTEST"));
