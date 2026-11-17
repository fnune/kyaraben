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
        devTools = with pkgs; [
          go_1_24
          gopls
          gotools
          go-tools
          golangci-lint
          delve
          nodejs_22
          nodePackages.npm
          nil
          nixpkgs-fmt
          git
          git-cliff
          jq
          just
          podman
          pre-commit
          prek
          xvfb-run
        ];
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

          site = pkgs.buildNpmPackage {
            pname = "kyaraben-site";
            version = "0.1.0";
            src = ./site;
            npmDepsHash = "sha256-aKmDWQk3mV4CCBo4MlqO95onAEvGwPHAw5ew6brcRzE=";
            installPhase = "cp -r dist $out";
          };

          relay = pkgs.buildGoModule {
            pname = "kyaraben-relay";
            version = "0.1.0";
            src = ./relay;
            vendorHash = null;
            subPackages = [ "cmd/relay" ];
            meta.mainProgram = "relay";
          };
        };

        devShells.ci = pkgs.mkShell {
          buildInputs = devTools;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = devTools ++ [ pkgs.musl ];

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
