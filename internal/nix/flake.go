package nix

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/fnune/kyaraben/internal/hardware"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/versions"
)

func retroArchCoreName(id model.EmulatorID) string {
	if !strings.HasPrefix(string(id), "retroarch:") {
		return ""
	}
	return strings.TrimPrefix(string(id), "retroarch:")
}

// EmulatorLookup provides access to emulator definitions.
type EmulatorLookup interface {
	GetEmulator(id model.EmulatorID) (model.Emulator, error)
}

// FrontendLookup provides access to frontend definitions.
type FrontendLookup interface {
	GetFrontend(id model.FrontendID) (model.Frontend, error)
}

type FlakeGenerator struct {
	emulators        EmulatorLookup
	frontends        FrontendLookup
	versionOverrides map[string]string // package name -> pinned version
}

func NewFlakeGenerator(emulators EmulatorLookup, frontends FrontendLookup) *FlakeGenerator {
	return &FlakeGenerator{
		emulators:        emulators,
		frontends:        frontends,
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
{{- if .Icons }}

        icons = pkgs.runCommand "kyaraben-icons" {} ''
          mkdir -p $out/share/icons
{{- range .Icons }}
          install -m644 ${pkgs.fetchurl {
            url = "{{ .URL }}";
            sha256 = "{{ .SHA256 }}";
          }} $out/share/icons/{{ .Filename }}
{{- end }}
        '';
{{- end }}
{{- if .RetroArchCores }}

        retroarch-cores = let
          coresBundle = pkgs.fetchurl {
            url = "{{ .RetroArchCoresURL }}";
            sha256 = "{{ .RetroArchCoresSHA256 }}";
          };
        in pkgs.runCommand "kyaraben-retroarch-cores" {} ''
          mkdir -p $out/lib/retroarch/cores
          ${pkgs.p7zip}/bin/7z x -o$TMPDIR ${coresBundle}
{{- range .RetroArchCoreFiles }}
          install -m644 $TMPDIR/RetroArch-Linux*/cores/{{ . }} $out/lib/retroarch/cores/
{{- end }}
        '';
{{- end }}

        default = pkgs.symlinkJoin {
          name = "kyaraben-profile";
          paths = [
{{- range .Launchers }}
            self.packages.${system}.{{ .Package }}
{{- end }}
{{- if .Icons }}
            self.packages.${system}.icons
{{- end }}
{{- if .RetroArchCores }}
            self.packages.${system}.retroarch-cores
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

type IconInfo struct {
	Name     string // emulator name (e.g., "eden")
	URL      string
	SHA256   string
	Filename string // e.g., "eden.png" or "eden.svg"
}

type GenerationPath string

type GenerateResult struct {
	Path             GenerationPath
	SkippedEmulators []model.EmulatorID
	SkippedFrontends []model.FrontendID
}

var retroArchCoreMapping = map[model.EmulatorID]string{
	model.EmulatorIDRetroArchBsnes:         "bsnes",
	model.EmulatorIDRetroArchMesen:         "mesen",
	model.EmulatorIDRetroArchGenesisPlusGX: "genesis-plus-gx",
	model.EmulatorIDRetroArchMupen64Plus:   "mupen64plus",
	model.EmulatorIDRetroArchBeetleSaturn:  "beetle-saturn",
}

func retroArchCorePackage(emuID model.EmulatorID) (string, bool) {
	pkg, ok := retroArchCoreMapping[emuID]
	return pkg, ok
}

func (fg *FlakeGenerator) Generate(baseDir string, emulatorIDs []model.EmulatorID, frontendIDs []model.FrontendID) (*GenerateResult, error) {
	if baseDir == "" {
		return nil, fmt.Errorf("baseDir cannot be empty")
	}
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	genDir := filepath.Join(baseDir, "generations", timestamp)

	if err := os.MkdirAll(genDir, 0755); err != nil {
		return nil, fmt.Errorf("creating generation directory: %w", err)
	}

	v := versions.MustGet()

	packages := make([]PackageInfo, 0, len(emulatorIDs)+len(frontendIDs))
	seenPackages := make(map[string]bool)
	seenBinaries := make(map[string]bool)
	launchers := make([]LauncherTemplateInfo, 0, len(emulatorIDs)+len(frontendIDs))
	icons := make([]IconInfo, 0, len(emulatorIDs)+len(frontendIDs))
	seenIcons := make(map[string]bool)
	retroArchCores := make([]string, 0)
	seenCores := make(map[string]bool)
	var skippedEmulators []model.EmulatorID
	var skippedFrontends []model.FrontendID

	for _, emuID := range emulatorIDs {
		emu, err := fg.emulators.GetEmulator(emuID)
		if err != nil {
			skippedEmulators = append(skippedEmulators, emuID)
			continue
		}

		if corePkg, ok := retroArchCorePackage(emuID); ok {
			if !seenCores[corePkg] {
				seenCores[corePkg] = true
				retroArchCores = append(retroArchCores, corePkg)
			}
		}

		pkg, err := fg.packageInfoFromRef(emu.Package)
		if err != nil {
			return nil, err
		}

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

		binaryName := emu.Launcher.Binary
		if binaryName != "" && !seenIcons[binaryName] {
			spec, ok := v.GetEmulator(pkg.Name)
			if ok && spec.IconURL != "" && spec.IconSHA256 != "" {
				seenIcons[binaryName] = true
				ext := filepath.Ext(spec.IconURL)
				icons = append(icons, IconInfo{
					Name:     binaryName,
					URL:      spec.IconURL,
					SHA256:   spec.IconSHA256,
					Filename: binaryName + ext,
				})
			}
		}
	}

	for _, feID := range frontendIDs {
		if fg.frontends == nil {
			skippedFrontends = append(skippedFrontends, feID)
			continue
		}
		fe, err := fg.frontends.GetFrontend(feID)
		if err != nil {
			skippedFrontends = append(skippedFrontends, feID)
			continue
		}

		pkg, err := fg.packageInfoFromRef(fe.Package)
		if err != nil {
			return nil, err
		}

		if !seenPackages[pkg.Name] {
			seenPackages[pkg.Name] = true
			packages = append(packages, pkg)
		}

		if fe.Launcher.Binary != "" && !seenBinaries[fe.Launcher.Binary] {
			seenBinaries[fe.Launcher.Binary] = true
			launchers = append(launchers, LauncherTemplateInfo{
				Package: pkg.Name,
			})
		}

		binaryName := fe.Launcher.Binary
		if binaryName != "" && !seenIcons[binaryName] {
			spec, ok := v.GetEmulator(pkg.Name)
			if ok && spec.IconURL != "" && spec.IconSHA256 != "" {
				seenIcons[binaryName] = true
				ext := filepath.Ext(spec.IconURL)
				icons = append(icons, IconInfo{
					Name:     binaryName,
					URL:      spec.IconURL,
					SHA256:   spec.IconSHA256,
					Filename: binaryName + ext,
				})
			}
		}
	}

	tmpl, err := template.New("flake").Parse(flakeTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	flakePath := filepath.Join(genDir, "flake.nix")
	f, err := os.Create(flakePath)
	if err != nil {
		return nil, fmt.Errorf("creating flake.nix: %w", err)
	}

	var retroArchCoresURL, retroArchCoresSHA256 string
	var retroArchCoreFiles []string
	if len(retroArchCores) > 0 {
		arch := "x86_64" // TODO: detect runtime arch
		url, sha256, ok := v.RetroArchCores.GetCoresURL(arch)
		if ok {
			retroArchCoresURL = url
			retroArchCoresSHA256 = sha256
			for _, coreName := range retroArchCores {
				if filename, ok := v.RetroArchCores.Files[coreName]; ok {
					retroArchCoreFiles = append(retroArchCoreFiles, filename)
				}
			}
		}
	}

	data := struct {
		NixpkgsCommit        string
		Packages             []PackageInfo
		Launchers            []LauncherTemplateInfo
		Icons                []IconInfo
		RetroArchCores       []string
		RetroArchCoresURL    string
		RetroArchCoresSHA256 string
		RetroArchCoreFiles   []string
	}{
		NixpkgsCommit:        v.Nixpkgs.Commit,
		Packages:             packages,
		Launchers:            launchers,
		Icons:                icons,
		RetroArchCores:       retroArchCores,
		RetroArchCoresURL:    retroArchCoresURL,
		RetroArchCoresSHA256: retroArchCoresSHA256,
		RetroArchCoreFiles:   retroArchCoreFiles,
	}

	execErr := tmpl.Execute(f, data)
	closeErr := f.Close()

	if execErr != nil {
		return nil, fmt.Errorf("executing template: %w", execErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("closing flake.nix: %w", closeErr)
	}

	return &GenerateResult{
		Path:             GenerationPath(genDir),
		SkippedEmulators: skippedEmulators,
		SkippedFrontends: skippedFrontends,
	}, nil
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

	if override, ok := fg.versionOverrides[name]; ok {
		entry := spec.GetVersion(override)
		if entry == nil {
			return nil, nil, fmt.Errorf("version %s not found for %s (available: %v)", override, name, spec.AvailableVersions())
		}
		return entry, spec, nil
	}

	entry := spec.GetDefault()
	if entry == nil {
		return nil, nil, fmt.Errorf("no default version for %s", name)
	}
	return entry, spec, nil
}

// packageInfoFromRef converts a PackageRef to nix package info.
func (fg *FlakeGenerator) packageInfoFromRef(ref model.PackageRef) (PackageInfo, error) {
	switch p := ref.(type) {
	case model.AppImage:
		return fg.packageInfoFromAppImage(p)
	default:
		return PackageInfo{}, fmt.Errorf("unsupported package source: %T", ref)
	}
}

// packageInfoFromAppImage generates nix expression for an AppImage from versions.toml
func (fg *FlakeGenerator) packageInfoFromAppImage(p model.AppImage) (PackageInfo, error) {
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

func packageInfoForDirectBinary(name string, entry *versions.VersionEntry, spec *versions.EmulatorSpec, x86Target, armTarget string, x86Build, armBuild *versions.TargetBuild) (PackageInfo, error) {
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

func (fg *FlakeGenerator) GetResolvedVersions(emulatorIDs []model.EmulatorID) map[model.EmulatorID]string {
	result := make(map[model.EmulatorID]string)
	for _, emuID := range emulatorIDs {
		emu, err := fg.emulators.GetEmulator(emuID)
		if err != nil {
			continue
		}
		entry, _, err := fg.getEmulatorVersion(emu.Package.PackageName())
		if err != nil {
			continue
		}
		result[emuID] = entry.Version
	}
	return result
}

func (fg *FlakeGenerator) GetResolvedFrontendVersions(frontendIDs []model.FrontendID) map[model.FrontendID]string {
	result := make(map[model.FrontendID]string)
	for _, feID := range frontendIDs {
		fe, err := fg.frontends.GetFrontend(feID)
		if err != nil {
			continue
		}
		entry, _, err := fg.getEmulatorVersion(fe.Package.PackageName())
		if err != nil {
			continue
		}
		result[feID] = entry.Version
	}
	return result
}

func (fg *FlakeGenerator) GetExpectedPackages(emulatorIDs []model.EmulatorID, frontendIDs []model.FrontendID) []ExpectedPackage {
	var packages []ExpectedPackage
	currentArch := getTargetTriple()

	vers, err := versions.Get()
	if err != nil {
		return packages
	}

	for _, emuID := range emulatorIDs {
		emu, err := fg.emulators.GetEmulator(emuID)
		if err != nil {
			continue
		}

		pkgName := emu.Package.PackageName()
		displayName := emu.Name
		var sizeBytes int64

		if spec, ok := vers.GetEmulator(pkgName); ok {
			if entry := spec.GetDefault(); entry != nil {
				if target := entry.DefaultTargetForArch(currentArch); target != "" {
					if build := entry.Target(target); build != nil && build.Size > 0 {
						sizeBytes = build.Size
					}
				}
			}
		}

		if coreName := retroArchCoreName(emuID); coreName != "" {
			coreLookupName := strings.ReplaceAll(coreName, "-", "_")
			coreSize := vers.GetCoreSize(coreLookupName)
			if coreSize > 0 {
				packages = append(packages, ExpectedPackage{
					Name:        "libretro-" + coreName,
					DisplayName: coreName + " (RetroArch)",
					SizeBytes:   coreSize,
				})
			}
		}

		packages = append(packages, ExpectedPackage{
			Name:        pkgName,
			DisplayName: displayName,
			SizeBytes:   sizeBytes,
		})
	}

	for _, feID := range frontendIDs {
		fe, err := fg.frontends.GetFrontend(feID)
		if err != nil {
			continue
		}

		pkgName := fe.Package.PackageName()
		displayName := fe.Name
		var sizeBytes int64

		if spec, ok := vers.GetEmulator(pkgName); ok {
			if entry := spec.GetDefault(); entry != nil {
				if target := entry.DefaultTargetForArch(currentArch); target != "" {
					if build := entry.Target(target); build != nil && build.Size > 0 {
						sizeBytes = build.Size
					}
				}
			}
		}

		packages = append(packages, ExpectedPackage{
			Name:        pkgName,
			DisplayName: displayName,
			SizeBytes:   sizeBytes,
		})
	}

	return packages
}
