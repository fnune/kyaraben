package launcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fnune/kyaraben/internal/logging"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/paths"
)

var log = logging.New("launcher")

type Manager struct {
	profileDir string
	dataDir    string
	resolver   model.BaseDirResolver
}

func NewManager() (*Manager, error) {
	return NewManagerWithResolver(model.OSBaseDirResolver{})
}

func NewManagerWithResolver(resolver model.BaseDirResolver) (*Manager, error) {
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
		resolver:   resolver,
	}, nil
}

func (m *Manager) ProfileDir() string {
	return m.profileDir
}

func (m *Manager) BinDir() string {
	return filepath.Join(m.profileDir, "bin")
}

type InstalledBinary struct {
	Name string
	Path string
}

type InstalledCore struct {
	Filename string
	Path     string
}

type InstalledIcon struct {
	Name     string
	Filename string
	Path     string
}

const wrapperTemplate = `#!/bin/sh
exec "{{.RealBinaryPath}}" "$@"
`

type wrapperData struct {
	RealBinaryPath string
}

func (m *Manager) GenerateWrappers(binaries []InstalledBinary) error {
	binDir := m.BinDir()

	if err := os.RemoveAll(binDir); err != nil {
		return fmt.Errorf("removing old bin directory: %w", err)
	}

	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("creating bin directory: %w", err)
	}

	if len(binaries) == 0 {
		return nil
	}

	tmpl, err := template.New("wrapper").Parse(wrapperTemplate)
	if err != nil {
		return fmt.Errorf("parsing wrapper template: %w", err)
	}

	for _, binary := range binaries {
		if strings.HasPrefix(binary.Name, ".") {
			continue
		}

		wrapperPath := filepath.Join(binDir, binary.Name)
		f, err := os.Create(wrapperPath)
		if err != nil {
			return fmt.Errorf("creating wrapper %s: %w", binary.Name, err)
		}

		data := wrapperData{RealBinaryPath: binary.Path}
		if err := tmpl.Execute(f, data); err != nil {
			_ = f.Close()
			return fmt.Errorf("writing wrapper %s: %w", binary.Name, err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("closing wrapper %s: %w", binary.Name, err)
		}

		if err := os.Chmod(wrapperPath, 0755); err != nil {
			return fmt.Errorf("making wrapper executable %s: %w", binary.Name, err)
		}

		log.Debug("Generated wrapper: %s -> %s", wrapperPath, binary.Path)
	}

	log.Info("Generated %d wrapper scripts in %s", len(binaries), binDir)
	return nil
}

func (m *Manager) GenerateCoreSymlinks(cores []InstalledCore) error {
	if len(cores) == 0 {
		return nil
	}

	coresDir, err := paths.RetroArchCoresDir()
	if err != nil {
		return fmt.Errorf("getting cores directory: %w", err)
	}

	if err := os.RemoveAll(coresDir); err != nil {
		return fmt.Errorf("removing old cores directory: %w", err)
	}

	if err := os.MkdirAll(coresDir, 0755); err != nil {
		return fmt.Errorf("creating cores directory: %w", err)
	}

	for _, core := range cores {
		destPath := filepath.Join(coresDir, core.Filename)
		if err := os.Symlink(core.Path, destPath); err != nil {
			return fmt.Errorf("creating symlink for %s: %w", core.Filename, err)
		}
		log.Debug("Created core symlink: %s -> %s", destPath, core.Path)
	}

	log.Info("Generated %d core symlinks in %s", len(cores), coresDir)
	return nil
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
	homeDir, err := m.resolver.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}
	dataDir, err := m.resolver.UserDataDir()
	if err != nil {
		return nil, fmt.Errorf("getting data directory: %w", err)
	}

	binDir := filepath.Join(homeDir, ".local", "bin")
	appsDir := filepath.Join(dataDir, "applications")

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

		if err := copyFile(currentExe, result.CLIPath); err != nil {
			return nil, fmt.Errorf("copying CLI: %w", err)
		}
		if err := os.Chmod(result.CLIPath, 0755); err != nil {
			return nil, fmt.Errorf("making CLI executable: %w", err)
		}
		log.Info("Installed CLI: %s (copied from %s)", result.CLIPath, currentExe)
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
	defer func() { _ = f.Close() }()

	if err := tmpl.Execute(f, struct{ ExecPath string }{ExecPath: execPath}); err != nil {
		return nil, fmt.Errorf("writing desktop file: %w", err)
	}
	log.Info("Installed desktop file: %s", result.DesktopPath)

	return result, nil
}

func (m *Manager) GetInstallStatus() *InstallResult {
	homeDir, err := m.resolver.UserHomeDir()
	if err != nil {
		return nil
	}
	dataDir, err := m.resolver.UserDataDir()
	if err != nil {
		return nil
	}

	binDir := filepath.Join(homeDir, ".local", "bin")
	appsDir := filepath.Join(dataDir, "applications")

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
