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
            go
            gopls
            gotools
            go-tools
            delve

            # Nix tools
            nil
            nixpkgs-fmt

            # General tools
            git
            jq
            podman
          ];

          shellHook = ''
            echo "Kyaraben development environment"
            echo "Go version: $(go version)"
          '';
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
