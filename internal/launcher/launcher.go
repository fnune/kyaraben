package launcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/logging"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/paths"
)

var log = logging.New("launcher")

type Manager struct {
	fs             vfs.FS
	paths          *paths.Paths
	profileDir     string
	dataDir        string
	resolver       model.BaseDirResolver
	executablePath string
}

func NewManager(p *paths.Paths) (*Manager, error) {
	stateDir, err := p.StateDir()
	if err != nil {
		return nil, fmt.Errorf("getting state directory: %w", err)
	}
	dataDir, err := paths.DataDir()
	if err != nil {
		return nil, fmt.Errorf("getting data directory: %w", err)
	}
	return &Manager{
		fs:         vfs.OSFS,
		paths:      p,
		profileDir: stateDir,
		dataDir:    dataDir,
		resolver:   model.OSBaseDirResolver{},
	}, nil
}

func NewDefaultManager() (*Manager, error) {
	return NewManager(paths.DefaultPaths())
}

func (m *Manager) CoresDir() string {
	return filepath.Join(m.profileDir, "cores")
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

const esdeWrapperTemplate = `#!/bin/sh
"{{.KyarabenPath}}" system esde-import 2>/dev/null || true
"{{.RealBinaryPath}}" "$@"
_exit_code=$?
"{{.KyarabenPath}}" system esde-export 2>/dev/null || true
exit $_exit_code
`

type wrapperData struct {
	RealBinaryPath string
	KyarabenPath   string
}

func (m *Manager) GenerateWrappers(binaries []InstalledBinary) error {
	binDir := m.BinDir()

	if err := m.fs.RemoveAll(binDir); err != nil {
		return fmt.Errorf("removing old bin directory: %w", err)
	}

	if err := vfs.MkdirAll(m.fs, binDir, 0755); err != nil {
		return fmt.Errorf("creating bin directory: %w", err)
	}

	if len(binaries) == 0 {
		return nil
	}

	tmpl, err := template.New("wrapper").Parse(wrapperTemplate)
	if err != nil {
		return fmt.Errorf("parsing wrapper template: %w", err)
	}

	esdeTmpl, err := template.New("esde-wrapper").Parse(esdeWrapperTemplate)
	if err != nil {
		return fmt.Errorf("parsing esde wrapper template: %w", err)
	}

	var kyarabenPath string
	if m.paths != nil {
		kyarabenPath, _ = m.paths.CLIInstallPath()
	}

	for _, binary := range binaries {
		if strings.HasPrefix(binary.Name, ".") {
			continue
		}

		wrapperPath := filepath.Join(binDir, binary.Name)
		f, err := m.fs.Create(wrapperPath)
		if err != nil {
			return fmt.Errorf("creating wrapper %s: %w", binary.Name, err)
		}

		data := wrapperData{RealBinaryPath: binary.Path, KyarabenPath: kyarabenPath}

		useTmpl := tmpl
		if binary.Name == "esde" {
			useTmpl = esdeTmpl
		}

		if err := useTmpl.Execute(f, data); err != nil {
			_ = f.Close()
			return fmt.Errorf("writing wrapper %s: %w", binary.Name, err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("closing wrapper %s: %w", binary.Name, err)
		}

		if err := m.fs.Chmod(wrapperPath, 0755); err != nil {
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

	coresDir := m.CoresDir()

	if err := m.fs.RemoveAll(coresDir); err != nil {
		return fmt.Errorf("removing old cores directory: %w", err)
	}

	if err := vfs.MkdirAll(m.fs, coresDir, 0755); err != nil {
		return fmt.Errorf("creating cores directory: %w", err)
	}

	for _, core := range cores {
		destPath := filepath.Join(coresDir, core.Filename)
		if err := m.fs.Symlink(core.Path, destPath); err != nil {
			return fmt.Errorf("creating symlink for %s: %w", core.Filename, err)
		}
		log.Debug("Created core symlink: %s -> %s", destPath, core.Path)
	}

	log.Info("Generated %d core symlinks in %s", len(cores), coresDir)
	return nil
}

const kyarabenDesktopTemplate = `[Desktop Entry]
Type=Application
Name=Kyaraben{{if .Instance}} ({{.Instance}}){{end}}
Comment=Declarative emulation manager
Exec={{.ExecPath}}{{if .Instance}} --instance {{.Instance}}{{end}}
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

	if err := vfs.MkdirAll(m.fs, binDir, 0755); err != nil {
		return nil, fmt.Errorf("creating bin directory: %w", err)
	}
	if err := vfs.MkdirAll(m.fs, appsDir, 0755); err != nil {
		return nil, fmt.Errorf("creating applications directory: %w", err)
	}

	result := &InstallResult{
		CLIPath:     filepath.Join(binDir, m.paths.CLIBinaryName()),
		DesktopPath: filepath.Join(appsDir, m.paths.DesktopFileName()),
	}

	if appImagePath != "" {
		result.AppPath = filepath.Join(binDir, m.paths.AppBinaryName())
		if err := m.copyFile(appImagePath, result.AppPath); err != nil {
			return nil, fmt.Errorf("copying AppImage: %w", err)
		}
		if err := m.fs.Chmod(result.AppPath, 0755); err != nil {
			return nil, fmt.Errorf("making AppImage executable: %w", err)
		}
		log.Info("Installed UI: %s", result.AppPath)
	}

	if _, err := m.fs.Lstat(result.CLIPath); err == nil {
		if err := m.fs.Remove(result.CLIPath); err != nil {
			return nil, fmt.Errorf("removing old CLI: %w", err)
		}
	}

	if sidecarPath != "" {
		if err := m.copyFile(sidecarPath, result.CLIPath); err != nil {
			return nil, fmt.Errorf("copying CLI: %w", err)
		}
		if err := m.fs.Chmod(result.CLIPath, 0755); err != nil {
			return nil, fmt.Errorf("making CLI executable: %w", err)
		}
		log.Info("Installed CLI: %s (copied from %s)", result.CLIPath, sidecarPath)
	} else {
		currentExe := m.executablePath
		if currentExe == "" {
			var err error
			currentExe, err = os.Executable()
			if err != nil {
				return nil, fmt.Errorf("getting current executable: %w", err)
			}
			currentExe, err = filepath.EvalSymlinks(currentExe)
			if err != nil {
				return nil, fmt.Errorf("resolving executable symlinks: %w", err)
			}
		}

		if err := m.copyFile(currentExe, result.CLIPath); err != nil {
			return nil, fmt.Errorf("copying CLI: %w", err)
		}
		if err := m.fs.Chmod(result.CLIPath, 0755); err != nil {
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

	f, err := m.fs.Create(result.DesktopPath)
	if err != nil {
		return nil, fmt.Errorf("creating desktop file: %w", err)
	}
	defer func() { _ = f.Close() }()

	templateData := struct {
		ExecPath string
		Instance string
	}{
		ExecPath: execPath,
		Instance: m.paths.Instance,
	}
	if err := tmpl.Execute(f, templateData); err != nil {
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

	appPath := filepath.Join(binDir, m.paths.AppBinaryName())
	if _, err := m.fs.Stat(appPath); err == nil {
		result.AppPath = appPath
	}

	cliPath := filepath.Join(binDir, m.paths.CLIBinaryName())
	if _, err := m.fs.Stat(cliPath); err == nil {
		result.CLIPath = cliPath
	}

	desktopPath := filepath.Join(appsDir, m.paths.DesktopFileName())
	if _, err := m.fs.Stat(desktopPath); err == nil {
		result.DesktopPath = desktopPath
	}

	return result
}

func (m *Manager) copyFile(src, dst string) error {
	data, err := m.fs.ReadFile(src)
	if err != nil {
		return err
	}
	return m.fs.WriteFile(dst, data, 0644)
}
