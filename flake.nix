{
  # Dev shell for running the Espresso e2e tests (.github/workflows/espresso-e2e-tests.yaml)
  # locally without mise:
  #
  #   nix develop            # or: nix develop -c <cmd>
  #   just build-contracts   # or, to skip forge fmt: cd packages/contracts-bedrock && just forge-build
  #   gotestsum --format testname -- -timeout 45m -p 1 -count 1 \
  #     -run 'TestBatcherSwitching|TestFallbackMechanismIntegrationTestChannelNotClosed|TestEspressoEnforcementHardfork|TestE2eDevnetWithEspressoDegradedLiveness' \
  #     ./espresso/environment/...
  #
  # Notes:
  # - In a git repo, flakes only see tracked files — `git add flake.nix flake.lock`
  #   is required for `nix develop` to work.
  # - forge is pinned to the official v1.2.3 release binaries (same version mise.toml
  #   pins for CI). Newer forge (e.g. 1.7.x) fails the contracts build: it applies
  #   `deny = "warnings"` to solc warnings that 1.2.3 tolerates.
  # - forge downloads the solc versions resolved from foundry.toml on first build
  #   (needs network).
  description = "Toolchain for celo-org/optimism Espresso e2e tests";

  inputs = {
    # Pinned nixos-unstable (2026-08): go 1.26.5, just 1.57.0, gotestsum 1.13.0.
    # mise.toml pins go 1.26.4; go 1.26.5 satisfies the go.mod toolchain directive.
    nixpkgs.url = "github:NixOS/nixpkgs/e72e4f299401a3689d4b3d5fc6496b11db7064eb";
  };

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});

      # Official foundry release binaries. The alpine builds are static musl
      # binaries, so they run unpatched from the nix store on any Linux.
      foundryVersion = "1.2.3";
      foundryAssets = {
        x86_64-linux = { asset = "alpine_amd64"; hash = "sha256-AVGAVxx5znZkcXdo2y816HjmrZiog3UFR4l0IMOixGE="; };
        aarch64-linux = { asset = "alpine_arm64"; hash = "sha256-xA9Hhr3jlFFMEHQ5Ckd7felXTnyVVhyksDsnXePxPBc="; };
        x86_64-darwin = { asset = "darwin_amd64"; hash = "sha256-4+K0JcfhuMhT7UVCdrIP9QD6gtZcyVrK0iW0twY63Uo="; };
        aarch64-darwin = { asset = "darwin_arm64"; hash = "sha256-o/PxQXp6AqFpQun8hkGNASHC2QTBE/6xsmydadxvKH0="; };
      };
      foundryFor = pkgs:
        let a = foundryAssets.${pkgs.stdenv.hostPlatform.system};
        in pkgs.stdenvNoCC.mkDerivation {
          pname = "foundry-bin";
          version = foundryVersion;
          src = pkgs.fetchurl {
            url = "https://github.com/foundry-rs/foundry/releases/download/v${foundryVersion}/foundry_v${foundryVersion}_${a.asset}.tar.gz";
            hash = a.hash;
          };
          sourceRoot = ".";
          dontBuild = true;
          installPhase = ''
            mkdir -p $out/bin
            install -m755 forge cast anvil chisel $out/bin/
          '';
        };
    in
    {
      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go_1_26
            just
            (foundryFor pkgs)
            gotestsum
            git
          ];
        };
      });
    };
}
