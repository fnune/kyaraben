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
        pkgs = nixpkgs.legacyPackages.${system};
        arch = if system == "x86_64-linux" then "x86_64" else "aarch64";
      });
    in {
      packages = forAllSystems ({ system, pkgs, arch }: {
{{- range .Packages }}
        {{ .Name }} = {{ .Expr }};
{{- end }}

        default = pkgs.runCommand "kyaraben-profile" {} ''
          mkdir -p $out/bin $out/share/applications $out/share/icons/hicolor/scalable/apps

{{- range .Launchers }}
          ln -s ${self.packages.${system}.{{ .Package }}}/bin/{{ .Binary }} $out/bin/{{ .Binary }}
{{- end }}

{{- range .Launchers }}
{{- if .HasDesktopFile }}
          if [ -d "${self.packages.${system}.{{ .Package }}}/share/icons" ]; then
            cp -rs ${self.packages.${system}.{{ .Package }}}/share/icons/* $out/share/icons/ 2>/dev/null || true
          fi
{{- end }}
{{- end }}

{{- range .Launchers }}
{{- if and (not .HasDesktopFile) .IconURL }}
          cp ${pkgs.fetchurl { url = "{{ .IconURL }}"; sha256 = "{{ .IconSHA256 }}"; }} $out/share/icons/hicolor/scalable/apps/{{ .Binary }}.svg
{{- end }}
{{- end }}

{{- range .Launchers }}
{{- if .HasDesktopFile }}
          for desktop in ${self.packages.${system}.{{ .Package }}}/share/applications/*.desktop; do
            if [ -f "$desktop" ]; then
              ln -s "$desktop" $out/share/applications/
            fi
          done
{{- else }}
          cat > $out/share/applications/{{ .Binary }}.desktop << 'DESKTOP'
[Desktop Entry]
Type=Application
Name={{ .Name }}
GenericName={{ .GenericName }}
Exec=$out/bin/{{ .Binary }}
Icon={{ .Binary }}
Terminal=false
Categories={{ .CategoriesStr }};
DESKTOP
          sed -i "s|\$out|$out|g" $out/share/applications/{{ .Binary }}.desktop
{{- end }}
{{- end }}
        '';
      });
    };
}
`

type PackageInfo struct {
	Name string
	Expr string
}

type LauncherTemplateInfo struct {
	Package        string // Nix package name
	Binary         string // Binary name
	Name           string // Display name
	GenericName    string // For .desktop GenericName
	CategoriesStr  string // Semicolon-separated categories
	HasDesktopFile bool   // Whether package has .desktop file
	IconURL        string // URL to fetch icon (for AppImages)
	IconSHA256     string // SHA256 of icon
}

// GenerationPath returns the path where flake.nix was generated.
type GenerationPath string

// Generate creates a flake.nix file in a timestamped generation subdirectory.
// Returns the generation directory path.
func (fg *FlakeGenerator) Generate(baseDir string, emulatorIDs []model.EmulatorID) (GenerationPath, error) {
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	genDir := filepath.Join(baseDir, "generations", timestamp)

	if err := os.MkdirAll(genDir, 0755); err != nil {
		return "", fmt.Errorf("creating generation directory: %w", err)
	}

	v := versions.MustGet()

	packages := make([]PackageInfo, 0, len(emulatorIDs))
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
		packages = append(packages, pkg)

		if emu.Launcher.Binary != "" && !seenBinaries[emu.Launcher.Binary] {
			seenBinaries[emu.Launcher.Binary] = true

			displayName := emu.Launcher.DisplayName
			if displayName == "" {
				displayName = emu.Name
			}

			hasDesktop := emu.Package.Source() == model.PackageSourceNixpkgs

			var iconURL, iconSHA256 string
			if p, ok := emu.Package.(model.VersionedAppImage); ok {
				if appimage := getAppImageVersion(p.Name); appimage != nil {
					iconURL = appimage.IconURL
					iconSHA256 = appimage.IconSHA256
				}
			}

			launchers = append(launchers, LauncherTemplateInfo{
				Package:        pkg.Name,
				Binary:         emu.Launcher.Binary,
				Name:           displayName,
				GenericName:    emu.Launcher.GenericName,
				CategoriesStr:  strings.Join(emu.Launcher.Categories, ";"),
				HasDesktopFile: hasDesktop,
				IconURL:        iconURL,
				IconSHA256:     iconSHA256,
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
        in pkgs.runCommand "%s-%s" {} ''
          mkdir -p $out/bin
          install -m755 ${src} $out/bin/%s
        ''`,
		appimage.URL(x86Target), x86Build.SHA256,
		appimage.URL(armTarget), armBuild.SHA256,
		p.Name, appimage.Version, p.Name,
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
