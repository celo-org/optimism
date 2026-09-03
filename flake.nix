{
  description = "Development shell for the OP Stack monorepo (Celo fork)";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forEachSystem = f:
        nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});

      # forge and abigen are the two tools whose version shows in the generated
      # contract bindings (op-service/bindings), so they are pinned to what CI
      # uses, read from mise.toml. nixpkgs ships newer versions of both. The
      # hashes below must be updated when the pins move.
      tools = (builtins.fromTOML (builtins.readFile ./mise.toml)).tools;
      foundryVersion = tools.forge;
      abigenVersion = tools."go:github.com/ethereum/go-ethereum/cmd/abigen";
    in
    {
      devShells = forEachSystem (pkgs:
        let
          platform = {
            x86_64-linux = "linux_amd64";
            aarch64-linux = "linux_arm64";
            x86_64-darwin = "darwin_amd64";
            aarch64-darwin = "darwin_arm64";
          }.${pkgs.stdenv.hostPlatform.system};

          # Official release binary.
          foundry = pkgs.stdenvNoCC.mkDerivation {
            pname = "foundry";
            version = foundryVersion;
            src = pkgs.fetchurl {
              url = "https://github.com/foundry-rs/foundry/releases/download/v${foundryVersion}/foundry_v${foundryVersion}_${platform}.tar.gz";
              hash = {
                linux_amd64 = "sha256-ggLzjxY1wnk7LRpP5EOub3MVGQ3G7tIZ15aaQKt4ooY=";
                linux_arm64 = "sha256-cGEv0dqd86izVIQhTk8rN7odbEEVCElJBkLXoFPDHqo=";
                darwin_amd64 = "sha256-4+K0JcfhuMhT7UVCdrIP9QD6gtZcyVrK0iW0twY63Uo=";
                darwin_arm64 = "sha256-o/PxQXp6AqFpQun8hkGNASHC2QTBE/6xsmydadxvKH0=";
              }.${platform};
            };
            sourceRoot = ".";
            nativeBuildInputs = pkgs.lib.optionals pkgs.stdenv.hostPlatform.isLinux [ pkgs.autoPatchelfHook ];
            buildInputs = pkgs.lib.optionals pkgs.stdenv.hostPlatform.isLinux [ pkgs.stdenv.cc.cc.lib ];
            installPhase = ''
              mkdir -p $out/bin
              install -m755 forge cast anvil chisel $out/bin/
            '';
          };

          # Built from the go-ethereum tag, as `go install` would.
          abigen = pkgs.buildGoModule {
            pname = "abigen";
            version = abigenVersion;
            src = pkgs.fetchFromGitHub {
              owner = "ethereum";
              repo = "go-ethereum";
              rev = "v${abigenVersion}";
              hash = "sha256-QVgi4IDPDGlvg8J3Wwrjtzkd6eYi3aAwU8IJbnIr5fU=";
            };
            vendorHash = "sha256-KRVI1DxjoABZFJkmjGaMVlmxIHvtSFuvmpuMuvr8Pws=";
            subPackages = [ "cmd/abigen" ];
            doCheck = false;
          };
        in
        {
          default = pkgs.mkShell {
            name = "optimism-dev";

            packages = with pkgs; [
              foundry
              abigen
              go
              just
              jq
              git
              # Shell linting, as CI runs it
              shellcheck
              shfmt
            ];

            # go.mod's go directive accepts any newer patch release; do not
            # download the exact one.
            GOTOOLCHAIN = "local";

            shellHook = ''
              printf '\n  OP Stack dev shell: forge %s, abigen %s, %s\n' \
                "$(forge --version | awk 'NR==1{print $3}' | cut -d- -f1)" \
                "$(abigen --version | awk '{print $3}')" \
                "$(go version | awk '{print $3}')"
              printf '  just gen-bindings / just check-bindings\n\n'
            '';
          };
        });
    };
}
