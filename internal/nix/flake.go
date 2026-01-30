package nix

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/fnune/kyaraben/internal/hardware"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/versions"
)

// EmulatorLookup provides access to emulator definitions.
type EmulatorLookup interface {
	GetEmulator(id model.EmulatorID) (model.Emulator, error)
}

type FlakeGenerator struct {
	emulators EmulatorLookup
}

func NewFlakeGenerator(emulators EmulatorLookup) *FlakeGenerator {
	return &FlakeGenerator{
		emulators: emulators,
	}
}

const flakeTemplate = `{
  description = "Kyaraben-managed emulator environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/{{ .NixpkgsCommit }}";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        arch = if system == "x86_64-linux" then "x86_64"
               else if system == "aarch64-linux" then "aarch64"
               else throw "Unsupported system: ${system}";
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
func (fg *FlakeGenerator) Generate(dir string, emulatorIDs []model.EmulatorID) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating flake directory: %w", err)
	}

	v := versions.MustGet()

	packages := make([]PackageInfo, 0, len(emulatorIDs))

	for _, emuID := range emulatorIDs {
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
		NixpkgsCommit string
		Packages      []PackageInfo
	}{
		NixpkgsCommit: v.Nixpkgs.Commit,
		Packages:      packages,
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
	emu, err := fg.emulators.GetEmulator(emuID)
	if err != nil {
		return PackageInfo{}, fmt.Errorf("unknown emulator: %s", emuID)
	}

	return packageInfoFromRef(emu.Package)
}

// packageInfoFromRef converts a PackageRef to nix package info.
func packageInfoFromRef(ref model.PackageRef) (PackageInfo, error) {
	switch p := ref.(type) {
	case model.NixpkgsPackage:
		expr := "pkgs." + p.Attr
		if p.Overlay != "" {
			expr = p.Overlay
		}
		return PackageInfo{
			Name: p.Attr,
			Expr: expr,
		}, nil

	case model.GitHubAppImage:
		return packageInfoFromGitHubAppImage(p), nil

	case model.VersionedAppImage:
		return packageInfoFromVersionedAppImage(p)

	default:
		return PackageInfo{}, fmt.Errorf("unsupported package source: %T", ref)
	}
}

// packageInfoFromGitHubAppImage generates nix expression for a GitHub AppImage.
// Deprecated: Use packageInfoFromVersionedAppImage instead
func packageInfoFromGitHubAppImage(p model.GitHubAppImage) PackageInfo {
	// Build a nix expression that selects the correct asset based on system architecture
	expr := fmt.Sprintf(`let
          assets = {
            x86_64 = {
              url = "https://github.com/%s/%s/releases/download/%s/%s";
              sha256 = "%s";
            };
            aarch64 = {
              url = "https://github.com/%s/%s/releases/download/%s/%s";
              sha256 = "%s";
            };
          };
          src = pkgs.fetchurl {
            url = assets.${arch}.url;
            sha256 = assets.${arch}.sha256;
          };
        in pkgs.appimageTools.wrapType2 {
          pname = "%s";
          version = "%s";
          src = src;
        }`,
		p.Owner, p.Repo, p.Version, p.Assets["x86_64"], p.Hashes["x86_64"],
		p.Owner, p.Repo, p.Version, p.Assets["aarch64"], p.Hashes["aarch64"],
		p.Name, p.Version,
	)
	return PackageInfo{
		Name: p.Name,
		Expr: expr,
	}
}

// packageInfoFromVersionedAppImage generates nix expression for an AppImage from versions.toml
func packageInfoFromVersionedAppImage(p model.VersionedAppImage) (PackageInfo, error) {
	v := versions.MustGet()

	var appimage *versions.AppImageVersion
	switch p.Name {
	case "eden":
		appimage = &v.Eden
	default:
		return PackageInfo{}, fmt.Errorf("unknown versioned appimage: %s", p.Name)
	}

	// Auto-detect hardware target for best performance
	detected := hardware.DetectTarget()
	x86Target := detected.Name
	if detected.Arch != "x86_64" {
		x86Target = appimage.DefaultTargetForArch("x86_64")
	}
	armTarget := "aarch64"

	x86Build := appimage.Target(x86Target)
	armBuild := appimage.Target(armTarget)
	if x86Build == nil || armBuild == nil {
		return PackageInfo{}, fmt.Errorf("missing target builds for %s", p.Name)
	}

	expr := fmt.Sprintf(`let
          assets = {
            x86_64 = {
              url = "%s";
              sha256 = "%s";
            };
            aarch64 = {
              url = "%s";
              sha256 = "%s";
            };
          };
          src = pkgs.fetchurl {
            url = assets.${arch}.url;
            sha256 = assets.${arch}.sha256;
          };
        in pkgs.appimageTools.wrapType2 {
          pname = "%s";
          version = "%s";
          src = src;
        }`,
		appimage.URL(x86Target), x86Build.SHA256,
		appimage.URL(armTarget), armBuild.SHA256,
		p.Name, appimage.Version,
	)

	return PackageInfo{
		Name: p.Name,
		Expr: expr,
	}, nil
}

// FlakeRef returns the flake reference for building an emulator.
func (fg *FlakeGenerator) FlakeRef(flakeDir string, emuID model.EmulatorID) (string, error) {
	emu, err := fg.emulators.GetEmulator(emuID)
	if err != nil {
		return "", fmt.Errorf("unknown emulator: %s", emuID)
	}

	pkg, err := packageInfoFromRef(emu.Package)
	if err != nil {
		return "", err
	}

	absPath, _ := filepath.Abs(flakeDir)
	return fmt.Sprintf("%s#%s", absPath, pkg.Name), nil
}

// DefaultFlakeRef returns the flake reference for the combined environment.
func (fg *FlakeGenerator) DefaultFlakeRef(flakeDir string) string {
	absPath, _ := filepath.Abs(flakeDir)
	return absPath
}
