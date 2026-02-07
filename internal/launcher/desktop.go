package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type GeneratedDesktop struct {
	BinaryName    string
	Name          string
	GenericName   string
	CategoriesStr string
	LaunchArgs    string // Command-line arguments to pass to the emulator
}

const desktopTemplate = `[Desktop Entry]
Type=Application
Name={{.Name}}
{{- if .GenericName}}
GenericName={{.GenericName}}
{{- end}}
Exec={{.BinDir}}/{{.BinaryName}}{{if .LaunchArgs}} {{.LaunchArgs}}{{end}} %f
Icon=kyaraben-{{.BinaryName}}
Categories={{.CategoriesStr}};
`

func (m *Manager) ApplicationsDir() string {
	return filepath.Join(m.dataDir, "applications", "kyaraben")
}

func (m *Manager) IconsDir() string {
	return filepath.Join(m.dataDir, "icons", "hicolor", "scalable", "apps")
}

func (m *Manager) iconsDirForExt(ext string) string {
	if ext == ".svg" {
		return filepath.Join(m.dataDir, "icons", "hicolor", "scalable", "apps")
	}
	return filepath.Join(m.dataDir, "icons", "hicolor", "256x256", "apps")
}

func (m *Manager) iconThemeDir() string {
	return filepath.Join(m.dataDir, "icons", "hicolor")
}

type GeneratedFiles struct {
	DesktopFiles []string
	IconFiles    []string
}

func (m *Manager) GenerateDesktopFiles(entries []GeneratedDesktop, previousFiles *GeneratedFiles) (*GeneratedFiles, error) {
	appsDir := m.ApplicationsDir()
	iconsDir := m.IconsDir()

	if previousFiles != nil {
		for _, f := range previousFiles.DesktopFiles {
			if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
				log.Debug("Failed to remove old desktop file %s: %v", f, err)
			}
		}
		for _, f := range previousFiles.IconFiles {
			if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
				log.Debug("Failed to remove old icon file %s: %v", f, err)
			}
		}
	}

	if err := os.MkdirAll(appsDir, 0755); err != nil {
		return nil, fmt.Errorf("creating applications directory: %w", err)
	}

	if err := os.MkdirAll(iconsDir, 0755); err != nil {
		return nil, fmt.Errorf("creating icons directory: %w", err)
	}

	tmpl, err := template.New("desktop").Parse(desktopTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing desktop template: %w", err)
	}

	result := &GeneratedFiles{}

	for _, entry := range entries {
		desktopPath, err := m.generateDesktopFile(tmpl, entry)
		if err != nil {
			return nil, fmt.Errorf("generating desktop file for %s: %w", entry.BinaryName, err)
		}
		result.DesktopFiles = append(result.DesktopFiles, desktopPath)

		iconPath, err := m.copyStoreIcon(entry.BinaryName)
		if err != nil {
			log.Debug("No icon for %s: %v", entry.BinaryName, err)
		} else {
			result.IconFiles = append(result.IconFiles, iconPath)
		}
	}

	m.updateIconCache()

	return result, nil
}

func (m *Manager) RemoveDesktopFiles(files *GeneratedFiles) {
	if files == nil {
		return
	}
	for _, f := range files.DesktopFiles {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			log.Debug("Failed to remove desktop file %s: %v", f, err)
		}
	}
	for _, f := range files.IconFiles {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			log.Debug("Failed to remove icon file %s: %v", f, err)
		}
	}
}

func (m *Manager) updateIconCache() []string {
	dataDir, _ := m.resolver.UserDataDir()
	applicationsDir := filepath.Join(dataDir, "applications")
	return UpdateIconCachesWithAppsDir(m.iconThemeDir(), applicationsDir)
}

func (m *Manager) RefreshIconCaches() []string {
	dataDir, _ := m.resolver.UserDataDir()
	applicationsDir := filepath.Join(dataDir, "applications")
	return UpdateIconCachesWithAppsDir(m.iconThemeDir(), applicationsDir)
}

func UpdateIconCaches(themeDir string) []string {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		homeDir, _ := os.UserHomeDir()
		dataDir = filepath.Join(homeDir, ".local", "share")
	}
	applicationsDir := filepath.Join(dataDir, "applications")
	return UpdateIconCachesWithAppsDir(themeDir, applicationsDir)
}

func UpdateIconCachesWithAppsDir(themeDir, applicationsDir string) []string {
	var refreshed []string

	// GTK-based DEs (GNOME, XFCE, etc.)
	if _, err := exec.LookPath("gtk-update-icon-cache"); err == nil {
		cmd := exec.Command("gtk-update-icon-cache", "-f", "-t", themeDir)
		if err := cmd.Run(); err != nil {
			log.Debug("gtk-update-icon-cache failed: %v", err)
		} else {
			refreshed = append(refreshed, "gtk-update-icon-cache")
		}
	}

	// Desktop database (freedesktop.org standard)
	if _, err := exec.LookPath("update-desktop-database"); err == nil {
		cmd := exec.Command("update-desktop-database", applicationsDir)
		if err := cmd.Run(); err != nil {
			log.Debug("update-desktop-database failed: %v", err)
		} else {
			refreshed = append(refreshed, "update-desktop-database")
		}
	}

	// KDE Plasma
	for _, kbuildsycoca := range []string{"kbuildsycoca6", "kbuildsycoca5"} {
		if _, err := exec.LookPath(kbuildsycoca); err == nil {
			cmd := exec.Command(kbuildsycoca)
			if err := cmd.Run(); err != nil {
				log.Debug("%s failed: %v", kbuildsycoca, err)
			} else {
				refreshed = append(refreshed, kbuildsycoca)
			}
			break
		}
	}

	return refreshed
}

func (m *Manager) virtualToRealStorePath(virtualPath string) string {
	const nixStorePrefix = "/nix/store/"
	if !strings.HasPrefix(virtualPath, nixStorePrefix) {
		return virtualPath
	}
	hashAndName := strings.TrimPrefix(virtualPath, nixStorePrefix)
	return filepath.Join(m.nixPortableLocation, ".nix-portable", "nix", "store", hashAndName)
}

type desktopTemplateData struct {
	BinaryName    string
	Name          string
	GenericName   string
	CategoriesStr string
	BinDir        string
	LaunchArgs    string
}

func (m *Manager) generateDesktopFile(tmpl *template.Template, entry GeneratedDesktop) (string, error) {
	desktopPath := filepath.Join(m.ApplicationsDir(), entry.BinaryName+".desktop")

	f, err := os.Create(desktopPath)
	if err != nil {
		return "", fmt.Errorf("creating desktop file: %w", err)
	}

	data := desktopTemplateData{
		BinaryName:    entry.BinaryName,
		Name:          entry.Name,
		GenericName:   entry.GenericName,
		CategoriesStr: entry.CategoriesStr,
		BinDir:        m.BinDir(),
		LaunchArgs:    entry.LaunchArgs,
	}

	execErr := tmpl.Execute(f, data)
	closeErr := f.Close()

	if execErr != nil {
		return "", fmt.Errorf("writing desktop file: %w", execErr)
	}
	if closeErr != nil {
		return "", fmt.Errorf("closing desktop file: %w", closeErr)
	}

	log.Debug("Generated desktop file: %s", desktopPath)
	return desktopPath, nil
}

// copyStoreIcon copies an icon from the nix store profile to the user's icons directory.
func (m *Manager) copyStoreIcon(binary string) (string, error) {
	storeIconsDir := filepath.Join(m.CurrentLink(), "share", "icons")

	entries, err := os.ReadDir(storeIconsDir)
	if err != nil {
		return "", fmt.Errorf("reading store icons dir: %w", err)
	}

	var symlinkPath string
	var ext string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, binary+".") {
			symlinkPath = filepath.Join(storeIconsDir, name)
			ext = filepath.Ext(name)
			break
		}
	}

	if symlinkPath == "" {
		return "", fmt.Errorf("icon not found for %s", binary)
	}

	// Icons in the profile are symlinks to /nix/store/... virtual paths.
	// On nix-portable systems, we need to convert to the real path.
	virtualTarget, err := os.Readlink(symlinkPath)
	if err != nil {
		return "", fmt.Errorf("reading icon symlink: %w", err)
	}
	realPath := m.virtualToRealStorePath(virtualTarget)

	iconData, err := os.ReadFile(realPath)
	if err != nil {
		return "", fmt.Errorf("reading source icon: %w", err)
	}

	destDir := m.iconsDirForExt(ext)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("creating icons directory: %w", err)
	}

	destPath := filepath.Join(destDir, "kyaraben-"+binary+ext)
	if err := os.WriteFile(destPath, iconData, 0644); err != nil {
		return "", fmt.Errorf("writing icon: %w", err)
	}

	log.Debug("Copied icon: %s -> %s", symlinkPath, destPath)
	return destPath, nil
}
