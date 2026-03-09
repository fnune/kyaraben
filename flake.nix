{
  description = "Kyaraben - Declarative emulation manager";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages = {
          default = self.packages.${system}.kyaraben;

          kyaraben = pkgs.buildGoModule {
            pname = "kyaraben";
            version = "0.1.0";
            src = ./.;
            vendorHash = null; # Will be set after first build

            meta = with pkgs.lib; {
              description = "Declarative emulation manager";
              homepage = "https://github.com/fnune/kyaraben";
              license = licenses.mit;
              maintainers = [ ];
            };
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go development
            go_1_24
            gopls
            gotools
            go-tools
            golangci-lint
            delve

            # Node.js development
            nodejs_22
            nodePackages.npm

            # Nix tools
            nil
            nixpkgs-fmt

            # Static linking for portable CGO binaries
            musl

            # General tools
            git
            git-cliff
            jq
            just
            podman
            pre-commit
            prek
            xvfb-run
          ];

          shellHook = ''
            echo "Kyaraben development environment"
            echo "Go version: $(go version)"
          '';

          CGO_ENABLED = "1";
          CC = "musl-gcc";
          GOFLAGS = "-tags=netgo";
          CGO_LDFLAGS = "-static -Wl,--no-as-needed";
        };

        # Emulator packages that kyaraben manages
        packages.emulators = {
          retroarch-bsnes = pkgs.retroarch.override {
            cores = with pkgs.libretro; [ bsnes ];
          };
          duckstation = pkgs.duckstation;
        };
      }
    );
}
