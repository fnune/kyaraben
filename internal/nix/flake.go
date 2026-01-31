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
	emulators EmulatorLookup
}

func NewFlakeGenerator(emulators EmulatorLookup) *FlakeGenerator {
	return &FlakeGenerator{
		emulators: emulators,
	}
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

		pkg, err := packageInfoFromRef(emu.Package)
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

func getAppImageVersion(name string) *versions.AppImageVersion {
	v := versions.MustGet()
	switch name {
	case "eden":
		return &v.Eden
	case "duckstation":
		return &v.DuckStation
	case "pcsx2":
		return &v.PCSX2
	case "ppsspp":
		return &v.PPSSPP
	case "mgba":
		return &v.MGBA
	case "cemu":
		return &v.Cemu
	case "azahar":
		return &v.Azahar
	case "dolphin":
		return &v.Dolphin
	case "melonds":
		return &v.MelonDS
	case "vita3k":
		return &v.Vita3K
	case "rpcs3":
		return &v.RPCS3
	case "flycast":
		return &v.Flycast
	case "retroarch":
		return &v.RetroArch
	case "tic80":
		return &v.TIC80
	default:
		return nil
	}
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
func packageInfoFromVersionedAppImage(p model.VersionedAppImage) (PackageInfo, error) {
	appimage := getAppImageVersion(p.Name)
	if appimage == nil {
		return PackageInfo{}, fmt.Errorf("unknown versioned appimage: %s", p.Name)
	}

	// Auto-detect hardware target for best performance
	detected := hardware.DetectTarget()
	x86Target := detected.Name
	if detected.Arch != "x86_64" || appimage.Target(x86Target) == nil {
		x86Target = appimage.DefaultTargetForArch("x86_64")
	}
	armTarget := appimage.DefaultTargetForArch("aarch64")

	x86Build := appimage.Target(x86Target)
	armBuild := appimage.Target(armTarget)

	// Handle single-arch packages (e.g., PCSX2 only has x86_64)
	if x86Build == nil && armBuild == nil {
		return PackageInfo{}, fmt.Errorf("no target builds for %s", p.Name)
	}

	archiveType := appimage.ArchiveType(x86Target)
	if archiveType != "" {
		return packageInfoForArchive(p.Name, appimage, x86Target, armTarget, x86Build, armBuild, archiveType)
	}

	return packageInfoForDirectBinary(p.Name, appimage, x86Target, armTarget, x86Build, armBuild)
}

// packageInfoForDirectBinary generates nix expression for direct binary downloads (AppImage, etc)
func packageInfoForDirectBinary(name string, appimage *versions.AppImageVersion, x86Target, armTarget string, x86Build, armBuild *versions.TargetBuild) (PackageInfo, error) {
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
          }`, appimage.URL(x86Target), x86Build.SHA256,
			appimage.URL(armTarget), armBuild.SHA256)
	} else if x86Build != nil {
		assetsExpr = fmt.Sprintf(`{
            x86_64 = {
              url = "%s";
              sha256 = "%s";
            };
          }`, appimage.URL(x86Target), x86Build.SHA256)
	} else {
		assetsExpr = fmt.Sprintf(`{
            aarch64 = {
              url = "%s";
              sha256 = "%s";
            };
          }`, appimage.URL(armTarget), armBuild.SHA256)
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
		name, appimage.Version, name,
	)

	return PackageInfo{
		Name: name,
		Expr: expr,
	}, nil
}

// packageInfoForArchive generates nix expression for archives that need extraction
func packageInfoForArchive(name string, appimage *versions.AppImageVersion, x86Target, armTarget string, x86Build, armBuild *versions.TargetBuild, archiveType string) (PackageInfo, error) {
	x86BinaryPath := appimage.BinaryPathForTarget(x86Target)
	armBinaryPath := appimage.BinaryPathForTarget(armTarget)

	if x86BinaryPath == "" {
		return PackageInfo{}, fmt.Errorf("binary_path required for archive package %s", name)
	}

	// Determine extraction command based on archive type
	var extractCmd string
	switch archiveType {
	case "7z":
		extractCmd = "${pkgs.p7zip}/bin/7z x -o$TMPDIR/extracted $src"
	case "tar.gz", "tgz":
		extractCmd = "${pkgs.gnutar}/bin/tar -xzf $src -C $TMPDIR/extracted"
	case "tar.xz":
		extractCmd = "${pkgs.gnutar}/bin/tar -xJf $src -C $TMPDIR/extracted"
	case "zip":
		extractCmd = "${pkgs.unzip}/bin/unzip -q $src -d $TMPDIR/extracted"
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
          }`, appimage.URL(x86Target), x86Build.SHA256, x86BinaryPath,
			appimage.URL(armTarget), armBuild.SHA256, armBinaryPath)
	} else if x86Build != nil {
		assetsExpr = fmt.Sprintf(`{
            x86_64 = {
              url = "%s";
              sha256 = "%s";
              binaryPath = "%s";
            };
          }`, appimage.URL(x86Target), x86Build.SHA256, x86BinaryPath)
	} else {
		assetsExpr = fmt.Sprintf(`{
            aarch64 = {
              url = "%s";
              sha256 = "%s";
              binaryPath = "%s";
            };
          }`, appimage.URL(armTarget), armBuild.SHA256, armBinaryPath)
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
		name, appimage.Version,
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

	pkg, err := packageInfoFromRef(emu.Package)
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
