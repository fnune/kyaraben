package nix

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/fnune/kyaraben/internal/hardware"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/versions"
)

// EmulatorLookup provides access to emulator definitions.
type EmulatorLookup interface {
	GetEmulator(id model.EmulatorID) (model.Emulator, error)
}

type FlakeGenerator struct {
	emulators        EmulatorLookup
	versionOverrides map[string]string // emulator name -> pinned version
}

func NewFlakeGenerator(emulators EmulatorLookup) *FlakeGenerator {
	return &FlakeGenerator{
		emulators:        emulators,
		versionOverrides: make(map[string]string),
	}
}

// SetVersionOverrides configures pinned versions from user config.
func (fg *FlakeGenerator) SetVersionOverrides(overrides map[string]string) {
	fg.versionOverrides = overrides
}

const flakeTemplate = `{
  description = "Kyaraben-managed emulator environment";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/{{ .NixpkgsCommit }}";

  outputs = { self, nixpkgs }:
    let
      forAllSystems = f: nixpkgs.lib.genAttrs [ "x86_64-linux" "aarch64-linux" ] (system: f {
        inherit system;
        pkgs = import nixpkgs {
          inherit system;
          config.allowUnfree = true;
        };
        arch = if system == "x86_64-linux" then "x86_64" else "aarch64";
      });
    in {
      packages = forAllSystems ({ system, pkgs, arch }: {
{{- range .Packages }}
        {{ .Name }} = {{ .Expr }};
{{- end }}

        default = pkgs.symlinkJoin {
          name = "kyaraben-profile";
          paths = [
{{- range .Launchers }}
            self.packages.${system}.{{ .Package }}
{{- end }}
          ];
        };
      });
    };
}
`

type PackageInfo struct {
	Name string
	Expr string
}

type LauncherTemplateInfo struct {
	Package string
}

type GenerationPath string

func (fg *FlakeGenerator) Generate(baseDir string, emulatorIDs []model.EmulatorID) (GenerationPath, error) {
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	genDir := filepath.Join(baseDir, "generations", timestamp)

	if err := os.MkdirAll(genDir, 0755); err != nil {
		return "", fmt.Errorf("creating generation directory: %w", err)
	}

	v := versions.MustGet()

	packages := make([]PackageInfo, 0, len(emulatorIDs))
	seenPackages := make(map[string]bool)
	seenBinaries := make(map[string]bool)
	launchers := make([]LauncherTemplateInfo, 0, len(emulatorIDs))

	for _, emuID := range emulatorIDs {
		emu, err := fg.emulators.GetEmulator(emuID)
		if err != nil {
			return "", fmt.Errorf("unknown emulator: %s", emuID)
		}

		pkg, err := fg.packageInfoFromRef(emu.Package)
		if err != nil {
			return "", err
		}

		// Only add package once even if multiple emulators share it
		if !seenPackages[pkg.Name] {
			seenPackages[pkg.Name] = true
			packages = append(packages, pkg)
		}

		if emu.Launcher.Binary != "" && !seenBinaries[emu.Launcher.Binary] {
			seenBinaries[emu.Launcher.Binary] = true
			launchers = append(launchers, LauncherTemplateInfo{
				Package: pkg.Name,
			})
		}
	}

	tmpl, err := template.New("flake").Parse(flakeTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	flakePath := filepath.Join(genDir, "flake.nix")
	f, err := os.Create(flakePath)
	if err != nil {
		return "", fmt.Errorf("creating flake.nix: %w", err)
	}

	data := struct {
		NixpkgsCommit string
		Packages      []PackageInfo
		Launchers     []LauncherTemplateInfo
	}{
		NixpkgsCommit: v.Nixpkgs.Commit,
		Packages:      packages,
		Launchers:     launchers,
	}

	execErr := tmpl.Execute(f, data)
	closeErr := f.Close()

	if execErr != nil {
		return "", fmt.Errorf("executing template: %w", execErr)
	}
	if closeErr != nil {
		return "", fmt.Errorf("closing flake.nix: %w", closeErr)
	}

	return GenerationPath(genDir), nil
}

// getEmulatorVersion returns the version entry and spec for an emulator,
// respecting any user-configured version overrides.
func (fg *FlakeGenerator) getEmulatorVersion(name string) (*versions.VersionEntry, *versions.EmulatorSpec, error) {
	v, err := versions.Get()
	if err != nil {
		return nil, nil, err
	}

	spec, ok := v.GetEmulator(name)
	if !ok {
		return nil, nil, fmt.Errorf("unknown emulator: %s", name)
	}

	// Check for user override
	if override, ok := fg.versionOverrides[name]; ok {
		entry := spec.GetVersion(override)
		if entry == nil {
			return nil, nil, fmt.Errorf("version %s not found for %s (available: %v)", override, name, spec.AvailableVersions())
		}
		return entry, spec, nil
	}

	// Use default
	entry := spec.GetDefault()
	if entry == nil {
		return nil, nil, fmt.Errorf("no default version for %s", name)
	}
	return entry, spec, nil
}

// packageInfoFromRef converts a PackageRef to nix package info.
func (fg *FlakeGenerator) packageInfoFromRef(ref model.PackageRef) (PackageInfo, error) {
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
		return fg.packageInfoFromVersionedAppImage(p)

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
        in pkgs.runCommand "%s-%s" {} ''
          mkdir -p $out/bin
          install -m755 ${src} $out/bin/%s
        ''`,
		p.Owner, p.Repo, p.Version, p.Assets["x86_64"], p.Hashes["x86_64"],
		p.Owner, p.Repo, p.Version, p.Assets["aarch64"], p.Hashes["aarch64"],
		p.Name, p.Version, p.Name,
	)
	return PackageInfo{
		Name: p.Name,
		Expr: expr,
	}
}

// packageInfoFromVersionedAppImage generates nix expression for an AppImage from versions.toml
func (fg *FlakeGenerator) packageInfoFromVersionedAppImage(p model.VersionedAppImage) (PackageInfo, error) {
	entry, spec, err := fg.getEmulatorVersion(p.Name)
	if err != nil {
		return PackageInfo{}, err
	}

	// Auto-detect hardware target for best performance
	detected := hardware.DetectTarget()
	x86Target := detected.Name
	if detected.Arch != "x86_64" || entry.Target(x86Target) == nil {
		x86Target = entry.DefaultTargetForArch("x86_64")
	}
	armTarget := entry.DefaultTargetForArch("aarch64")

	x86Build := entry.Target(x86Target)
	armBuild := entry.Target(armTarget)

	// Handle single-arch packages (e.g., PCSX2 only has x86_64)
	if x86Build == nil && armBuild == nil {
		return PackageInfo{}, fmt.Errorf("no target builds for %s", p.Name)
	}

	archiveType := entry.ArchiveType(x86Target, spec)
	if archiveType != "" {
		return packageInfoForArchive(p.Name, entry, spec, x86Target, armTarget, x86Build, armBuild, archiveType)
	}

	return packageInfoForDirectBinary(p.Name, entry, spec, x86Target, armTarget, x86Build, armBuild)
}

// packageInfoForDirectBinary generates nix expression for direct binary downloads (AppImage, etc)
func packageInfoForDirectBinary(name string, entry *versions.VersionEntry, spec *versions.EmulatorSpec, x86Target, armTarget string, x86Build, armBuild *versions.TargetBuild) (PackageInfo, error) {
	// Build assets map, handling single-arch packages
	var assetsExpr string
	if x86Build != nil && armBuild != nil {
		assetsExpr = fmt.Sprintf(`{
            x86_64 = {
              url = "%s";
              sha256 = "%s";
            };
            aarch64 = {
              url = "%s";
              sha256 = "%s";
            };
          }`, entry.URL(x86Target, spec), x86Build.SHA256,
			entry.URL(armTarget, spec), armBuild.SHA256)
	} else if x86Build != nil {
		assetsExpr = fmt.Sprintf(`{
            x86_64 = {
              url = "%s";
              sha256 = "%s";
            };
          }`, entry.URL(x86Target, spec), x86Build.SHA256)
	} else {
		assetsExpr = fmt.Sprintf(`{
            aarch64 = {
              url = "%s";
              sha256 = "%s";
            };
          }`, entry.URL(armTarget, spec), armBuild.SHA256)
	}

	expr := fmt.Sprintf(`let
          assets = %s;
          src = pkgs.fetchurl {
            url = assets.${arch}.url;
            sha256 = assets.${arch}.sha256;
          };
        in pkgs.runCommand "%s-%s" {} ''
          mkdir -p $out/bin
          install -m755 ${src} $out/bin/%s
        ''`,
		assetsExpr,
		name, entry.Version, name,
	)

	return PackageInfo{
		Name: name,
		Expr: expr,
	}, nil
}

// packageInfoForArchive generates nix expression for archives that need extraction
func packageInfoForArchive(name string, entry *versions.VersionEntry, spec *versions.EmulatorSpec, x86Target, armTarget string, x86Build, armBuild *versions.TargetBuild, archiveType string) (PackageInfo, error) {
	x86BinaryPath := entry.BinaryPathForTarget(x86Target, spec)
	armBinaryPath := entry.BinaryPathForTarget(armTarget, spec)

	if x86BinaryPath == "" {
		return PackageInfo{}, fmt.Errorf("binary_path required for archive package %s", name)
	}

	// Determine extraction command based on archive type
	// Note: ${src} uses Nix interpolation to reference the let-bound src
	var extractCmd string
	switch archiveType {
	case "7z":
		extractCmd = "${pkgs.p7zip}/bin/7z x -o$TMPDIR/extracted ${src}"
	case "tar.gz", "tgz":
		extractCmd = "${pkgs.gnutar}/bin/tar -xzf ${src} -C $TMPDIR/extracted"
	case "tar.xz":
		extractCmd = "${pkgs.gnutar}/bin/tar -xJf ${src} -C $TMPDIR/extracted"
	case "zip":
		extractCmd = "${pkgs.unzip}/bin/unzip -q ${src} -d $TMPDIR/extracted"
	default:
		return PackageInfo{}, fmt.Errorf("unsupported archive type: %s", archiveType)
	}

	// Build assets map with binary paths
	var assetsExpr string
	if x86Build != nil && armBuild != nil {
		if armBinaryPath == "" {
			armBinaryPath = x86BinaryPath
		}
		assetsExpr = fmt.Sprintf(`{
            x86_64 = {
              url = "%s";
              sha256 = "%s";
              binaryPath = "%s";
            };
            aarch64 = {
              url = "%s";
              sha256 = "%s";
              binaryPath = "%s";
            };
          }`, entry.URL(x86Target, spec), x86Build.SHA256, x86BinaryPath,
			entry.URL(armTarget, spec), armBuild.SHA256, armBinaryPath)
	} else if x86Build != nil {
		assetsExpr = fmt.Sprintf(`{
            x86_64 = {
              url = "%s";
              sha256 = "%s";
              binaryPath = "%s";
            };
          }`, entry.URL(x86Target, spec), x86Build.SHA256, x86BinaryPath)
	} else {
		assetsExpr = fmt.Sprintf(`{
            aarch64 = {
              url = "%s";
              sha256 = "%s";
              binaryPath = "%s";
            };
          }`, entry.URL(armTarget, spec), armBuild.SHA256, armBinaryPath)
	}

	expr := fmt.Sprintf(`let
          assets = %s;
          src = pkgs.fetchurl {
            url = assets.${arch}.url;
            sha256 = assets.${arch}.sha256;
          };
        in pkgs.runCommand "%s-%s" {} ''
          mkdir -p $TMPDIR/extracted
          %s
          mkdir -p $out/bin
          install -m755 $TMPDIR/extracted/${assets.${arch}.binaryPath} $out/bin/%s
        ''`,
		assetsExpr,
		name, entry.Version,
		extractCmd,
		name,
	)

	return PackageInfo{
		Name: name,
		Expr: expr,
	}, nil
}

func (fg *FlakeGenerator) FlakeRef(flakeDir string, emuID model.EmulatorID) (string, error) {
	emu, err := fg.emulators.GetEmulator(emuID)
	if err != nil {
		return "", fmt.Errorf("unknown emulator: %s", emuID)
	}

	pkg, err := fg.packageInfoFromRef(emu.Package)
	if err != nil {
		return "", err
	}

	absPath, _ := filepath.Abs(flakeDir)
	return fmt.Sprintf("%s#%s", absPath, pkg.Name), nil
}

func (fg *FlakeGenerator) DefaultFlakeRef(flakeDir string) string {
	absPath, _ := filepath.Abs(flakeDir)
	return absPath
}
