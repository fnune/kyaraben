package launcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fnune/kyaraben/internal/logging"
	"github.com/fnune/kyaraben/internal/paths"
)

var log = logging.New("launcher")

type Manager struct {
	profileDir          string
	dataDir             string
	nixPortableBinary   string
	nixPortableLocation string
}

func NewManager() (*Manager, error) {
	stateDir, err := paths.KyarabenStateDir()
	if err != nil {
		return nil, fmt.Errorf("getting state directory: %w", err)
	}
	dataDir, err := paths.DataDir()
	if err != nil {
		return nil, fmt.Errorf("getting data directory: %w", err)
	}
	return &Manager{
		profileDir: stateDir,
		dataDir:    dataDir,
	}, nil
}

func (m *Manager) SetNixPortableBinary(path string) {
	m.nixPortableBinary = path
}

func (m *Manager) SetNixPortableLocation(path string) {
	m.nixPortableLocation = path
}

func (m *Manager) ProfileDir() string {
	return m.profileDir
}

func (m *Manager) CurrentLink() string {
	return filepath.Join(m.profileDir, "current")
}

func (m *Manager) Link(storePath string) error {
	log.Info("Linking profile: %s -> %s", m.CurrentLink(), storePath)

	if err := os.MkdirAll(m.profileDir, 0755); err != nil {
		return fmt.Errorf("creating profile directory: %w", err)
	}

	currentLink := m.CurrentLink()

	if _, err := os.Lstat(currentLink); err == nil {
		log.Debug("Removing old symlink: %s", currentLink)
		if err := os.Remove(currentLink); err != nil {
			return fmt.Errorf("removing old symlink: %w", err)
		}
	}

	if err := os.Symlink(storePath, currentLink); err != nil {
		return fmt.Errorf("creating symlink: %w", err)
	}

	if target, err := os.Readlink(currentLink); err == nil {
		log.Info("Created symlink: %s -> %s", currentLink, target)
		if entries, err := os.ReadDir(filepath.Join(target, "bin")); err == nil {
			binaries := make([]string, 0, len(entries))
			for _, e := range entries {
				binaries = append(binaries, e.Name())
			}
			log.Info("Available binaries: %v", binaries)
		}
	}

	return nil
}

func (m *Manager) Unlink() error {
	currentLink := m.CurrentLink()
	if _, err := os.Lstat(currentLink); err == nil {
		return os.Remove(currentLink)
	}
	return nil
}

func (m *Manager) BinDir() string {
	return filepath.Join(m.profileDir, "bin")
}

type EmulatorPackageInfo struct {
	BinaryName string
}

const wrapperTemplate = `#!/bin/sh
exec "{{.RealBinaryPath}}" "$@"
`

type wrapperData struct {
	RealBinaryPath string
}

func (m *Manager) GenerateWrappers(emulators []EmulatorPackageInfo) error {
	if m.nixPortableBinary == "" {
		return fmt.Errorf("nix-portable binary path not set")
	}

	binDir := m.BinDir()
	profileBinDir := filepath.Join(m.CurrentLink(), "bin")

	if _, err := os.Stat(profileBinDir); os.IsNotExist(err) {
		log.Info("No bin directory in profile, skipping wrapper generation")
		return nil
	}

	if err := os.RemoveAll(binDir); err != nil {
		return fmt.Errorf("removing old bin directory: %w", err)
	}

	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("creating bin directory: %w", err)
	}

	entries, err := os.ReadDir(profileBinDir)
	if err != nil {
		return fmt.Errorf("reading profile bin directory: %w", err)
	}

	tmpl, err := template.New("wrapper").Parse(wrapperTemplate)
	if err != nil {
		return fmt.Errorf("parsing wrapper template: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		binaryName := entry.Name()

		if strings.HasPrefix(binaryName, ".") {
			continue
		}
		wrapperPath := filepath.Join(binDir, binaryName)

		f, err := os.Create(wrapperPath)
		if err != nil {
			return fmt.Errorf("creating wrapper %s: %w", binaryName, err)
		}

		symlinkPath := filepath.Join(profileBinDir, binaryName)
		virtualTarget, err := os.Readlink(symlinkPath)
		if err != nil {
			_ = f.Close()
			return fmt.Errorf("reading symlink for %s: %w", binaryName, err)
		}
		realBinaryPath := m.virtualToRealStorePath(virtualTarget)
		data := wrapperData{RealBinaryPath: realBinaryPath}
		if err := tmpl.Execute(f, data); err != nil {
			_ = f.Close()
			return fmt.Errorf("writing wrapper %s: %w", binaryName, err)
		}
		log.Debug("Generated wrapper: %s -> %s", wrapperPath, realBinaryPath)

		if err := f.Close(); err != nil {
			return fmt.Errorf("closing wrapper %s: %w", binaryName, err)
		}

		if err := os.Chmod(wrapperPath, 0755); err != nil {
			return fmt.Errorf("making wrapper executable %s: %w", binaryName, err)
		}
	}

	log.Info("Generated %d wrapper scripts in %s", len(entries), binDir)
	return nil
}

func realToVirtualStorePath(realPath string) (string, error) {
	const storeMarker = "/nix/store/"
	idx := strings.Index(realPath, storeMarker)
	if idx == -1 {
		return "", fmt.Errorf("path does not contain /nix/store/: %s", realPath)
	}
	return realPath[idx:], nil
}

const kyarabenDesktopTemplate = `[Desktop Entry]
Type=Application
Name=Kyaraben
Comment=Declarative emulation manager
Exec={{.ExecPath}}
Icon=applications-games
Terminal=false
Categories=Game;Emulator;
`

type InstallResult struct {
	AppPath     string
	DesktopPath string
	CLIPath     string
}

func (m *Manager) InstallKyaraben(appImagePath, sidecarPath string) (*InstallResult, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}

	binDir := filepath.Join(homeDir, ".local", "bin")
	appsDir := filepath.Join(homeDir, ".local", "share", "applications")

	if err := os.MkdirAll(binDir, 0755); err != nil {
		return nil, fmt.Errorf("creating bin directory: %w", err)
	}
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		return nil, fmt.Errorf("creating applications directory: %w", err)
	}

	result := &InstallResult{
		CLIPath:     filepath.Join(binDir, "kyaraben"),
		DesktopPath: filepath.Join(appsDir, "kyaraben.desktop"),
	}

	if appImagePath != "" {
		result.AppPath = filepath.Join(binDir, "kyaraben-ui")
		if err := copyFile(appImagePath, result.AppPath); err != nil {
			return nil, fmt.Errorf("copying AppImage: %w", err)
		}
		if err := os.Chmod(result.AppPath, 0755); err != nil {
			return nil, fmt.Errorf("making AppImage executable: %w", err)
		}
		log.Info("Installed UI: %s", result.AppPath)
	}

	if _, err := os.Lstat(result.CLIPath); err == nil {
		if err := os.Remove(result.CLIPath); err != nil {
			return nil, fmt.Errorf("removing old CLI: %w", err)
		}
	}

	if sidecarPath != "" {
		if err := copyFile(sidecarPath, result.CLIPath); err != nil {
			return nil, fmt.Errorf("copying CLI: %w", err)
		}
		if err := os.Chmod(result.CLIPath, 0755); err != nil {
			return nil, fmt.Errorf("making CLI executable: %w", err)
		}
		log.Info("Installed CLI: %s (copied from %s)", result.CLIPath, sidecarPath)
	} else {
		currentExe, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("getting current executable: %w", err)
		}
		currentExe, err = filepath.EvalSymlinks(currentExe)
		if err != nil {
			return nil, fmt.Errorf("resolving executable symlinks: %w", err)
		}

		if err := os.Symlink(currentExe, result.CLIPath); err != nil {
			return nil, fmt.Errorf("creating CLI symlink: %w", err)
		}
		log.Info("Installed CLI: %s -> %s", result.CLIPath, currentExe)
	}

	tmpl, err := template.New("desktop").Parse(kyarabenDesktopTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing desktop template: %w", err)
	}

	execPath := result.CLIPath
	if result.AppPath != "" {
		execPath = result.AppPath
	}

	f, err := os.Create(result.DesktopPath)
	if err != nil {
		return nil, fmt.Errorf("creating desktop file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, struct{ ExecPath string }{ExecPath: execPath}); err != nil {
		return nil, fmt.Errorf("writing desktop file: %w", err)
	}
	log.Info("Installed desktop file: %s", result.DesktopPath)

	return result, nil
}

func (m *Manager) GetInstallStatus() *InstallResult {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	binDir := filepath.Join(homeDir, ".local", "bin")
	appsDir := filepath.Join(homeDir, ".local", "share", "applications")

	result := &InstallResult{}

	appPath := filepath.Join(binDir, "kyaraben-ui")
	if _, err := os.Stat(appPath); err == nil {
		result.AppPath = appPath
	}

	cliPath := filepath.Join(binDir, "kyaraben")
	if _, err := os.Stat(cliPath); err == nil {
		result.CLIPath = cliPath
	}

	desktopPath := filepath.Join(appsDir, "kyaraben.desktop")
	if _, err := os.Stat(desktopPath); err == nil {
		result.DesktopPath = desktopPath
	}

	return result
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
