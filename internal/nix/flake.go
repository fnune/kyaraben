package nix

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fnune/kyaraben/internal/model"
)

type FlakeGenerator struct {
	emulatorAttrs map[model.EmulatorID]string
}

func NewFlakeGenerator() *FlakeGenerator {
	return &FlakeGenerator{
		emulatorAttrs: map[model.EmulatorID]string{
			model.EmulatorRetroArchBsnes: "retroarch-bsnes",
			model.EmulatorDuckStation:    "duckstation",
			model.EmulatorTIC80:          "tic-80",
			model.EmulatorE2ETest:        "hello",
		},
	}
}

const flakeTemplate = `{
  description = "Kyaraben-managed emulator environment";

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
{{- range .Packages }}
          {{ .Name }} = {{ .Expr }};
{{- end }}

          # Combined environment with all emulators
          default = pkgs.symlinkJoin {
            name = "kyaraben-emulators";
            paths = [
{{- range .Packages }}
              self.packages.${system}.{{ .Name }}
{{- end }}
            ];
          };
        };
      }
    );
}
`

type PackageInfo struct {
	Name string
	Expr string
}

// Generate creates a flake.nix file in the given directory.
func (fg *FlakeGenerator) Generate(dir string, emulators []model.EmulatorID) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating flake directory: %w", err)
	}

	packages := make([]PackageInfo, 0, len(emulators))

	for _, emuID := range emulators {
		pkg, err := fg.packageForEmulator(emuID)
		if err != nil {
			return err
		}
		packages = append(packages, pkg)
	}

	tmpl, err := template.New("flake").Parse(flakeTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	flakePath := filepath.Join(dir, "flake.nix")
	f, err := os.Create(flakePath)
	if err != nil {
		return fmt.Errorf("creating flake.nix: %w", err)
	}

	data := struct {
		Packages []PackageInfo
	}{
		Packages: packages,
	}

	execErr := tmpl.Execute(f, data)
	closeErr := f.Close()

	if execErr != nil {
		return fmt.Errorf("executing template: %w", execErr)
	}
	if closeErr != nil {
		return fmt.Errorf("closing flake.nix: %w", closeErr)
	}

	return nil
}

func (fg *FlakeGenerator) packageForEmulator(emuID model.EmulatorID) (PackageInfo, error) {
	// Handle special cases
	switch emuID {
	case model.EmulatorRetroArchBsnes:
		return PackageInfo{
			Name: "retroarch-bsnes",
			Expr: `pkgs.retroarch.override { cores = with pkgs.libretro; [ bsnes ]; }`,
		}, nil

	case model.EmulatorDuckStation:
		return PackageInfo{
			Name: "duckstation",
			Expr: "pkgs.duckstation",
		}, nil

	case model.EmulatorTIC80:
		return PackageInfo{
			Name: "tic-80",
			Expr: "pkgs.tic-80",
		}, nil

	case model.EmulatorE2ETest:
		return PackageInfo{
			Name: "hello",
			Expr: "pkgs.hello",
		}, nil

	default:
		return PackageInfo{}, fmt.Errorf("unknown emulator: %s", emuID)
	}
}

// FlakeRef returns the flake reference for building an emulator.
func (fg *FlakeGenerator) FlakeRef(flakeDir string, emuID model.EmulatorID) string {
	attr := fg.emulatorAttrs[emuID]
	if attr == "" {
		attr = string(emuID)
	}
	// Normalize the path for nix
	absPath, _ := filepath.Abs(flakeDir)
	return fmt.Sprintf("%s#%s", absPath, strings.ReplaceAll(attr, ":", "-"))
}

// DefaultFlakeRef returns the flake reference for the combined environment.
func (fg *FlakeGenerator) DefaultFlakeRef(flakeDir string) string {
	absPath, _ := filepath.Abs(flakeDir)
	return absPath
}
