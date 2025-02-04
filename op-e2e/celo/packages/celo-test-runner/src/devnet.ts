import type { Config } from "./types.ts";

export async function setupDevnet(config: Config) {
  if (!config.SpawnDevnet) {
    return;
  }
  console.log("config requested to spawn devnet ...");

  const env = {
    DEVNET_L2OO: !config.UseFaultproofSystem,
    GENERIC_ALTDA: config.UseAltDA,
    DEVNET_ALTDA: config.UseAltDA,
    DEVNET_CELO: true,
  };

  const envAsString: Record<string, string> = Object.fromEntries(
    Object.entries(env).map(([key, value]) => [key, String(value)]),
  );
  const devnetUp = new Deno.Command("make", {
    args: ["devnet-up"],
    stdout: "piped",
    stderr: "piped",
    cwd: config.MonorepoPath,
    env: envAsString,
  });

  const process = devnetUp.spawn();
  const { code, stderr } = await process.output();

  if (code !== 0) {
    const errorOutput = new TextDecoder().decode(stderr);
    console.error("failed to spawn devnet:\n", errorOutput);
    Deno.exit(1);
  }

  console.log("devnet started.");
}

export async function teardownDevnet(config: Config) {
  if (!config.SpawnDevnet) {
    return;
  }
  console.log("stopping and clearing devnet...");
  const devnetClean = new Deno.Command("make", {
    args: ["devnet-clean"],
    stdout: "piped",
    stderr: "piped",
    cwd: config.MonorepoPath,
  });

  const process = devnetClean.spawn();
  const { code, stderr } = await process.output();

  if (code !== 0) {
    const errorOutput = new TextDecoder().decode(stderr);
    console.error("Failed to clean devnet:\n", errorOutput);
  }
  console.log("cleaned devnet.");
}
