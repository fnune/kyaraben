package launcher

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/fnune/kyaraben/internal/paths"
)

//go:embed icons/*.svg
var embeddedIcons embed.FS

type DesktopEntry interface {
	Binary() string
	isDesktopEntry()
}

type NixStoreDesktop struct {
	BinaryName string
}

func (n NixStoreDesktop) Binary() string  { return n.BinaryName }
func (n NixStoreDesktop) isDesktopEntry() {}

type GeneratedDesktop struct {
	BinaryName    string
	Name          string
	GenericName   string
	CategoriesStr string
}

func (g GeneratedDesktop) Binary() string  { return g.BinaryName }
func (g GeneratedDesktop) isDesktopEntry() {}

const desktopTemplate = `[Desktop Entry]
Type=Application
Name={{.Name}}
{{- if .GenericName}}
GenericName={{.GenericName}}
{{- end}}
Exec={{.BinDir}}/{{.BinaryName}} %f
Icon={{.BinaryName}}
Categories={{.CategoriesStr}};
`

func (m *Manager) ApplicationsDir() string {
	dataDir, _ := paths.DataDir()
	return filepath.Join(dataDir, "applications")
}

func (m *Manager) IconsDir() string {
	dataDir, _ := paths.DataDir()
	return filepath.Join(dataDir, "icons", "hicolor", "scalable", "apps")
}

func (m *Manager) iconThemeDir() string {
	dataDir, _ := paths.DataDir()
	return filepath.Join(dataDir, "icons", "hicolor")
}

type GeneratedFiles struct {
	DesktopFiles []string
	IconFiles    []string
}

func (m *Manager) GenerateDesktopFiles(entries []DesktopEntry, previousFiles *GeneratedFiles) (*GeneratedFiles, error) {
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
		switch e := entry.(type) {
		case NixStoreDesktop:
			files, err := m.copyNixStoreDesktop(e.BinaryName)
			if err != nil {
				log.Debug("Failed to copy desktop file for %s: %v", e.BinaryName, err)
			} else {
				result.DesktopFiles = append(result.DesktopFiles, files...)
			}
		case GeneratedDesktop:
			desktopPath, err := m.generateDesktopFile(tmpl, e)
			if err != nil {
				return nil, fmt.Errorf("generating desktop file for %s: %w", e.BinaryName, err)
			}
			result.DesktopFiles = append(result.DesktopFiles, desktopPath)

			iconPath, err := m.writeEmbeddedIcon(e.BinaryName)
			if err != nil {
				log.Debug("No embedded icon for %s: %v", e.BinaryName, err)
			} else {
				result.IconFiles = append(result.IconFiles, iconPath)
			}
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

func (m *Manager) updateIconCache() {
	themeDir := m.iconThemeDir()

	// GTK-based DEs (GNOME, XFCE, etc.)
	if _, err := exec.LookPath("gtk-update-icon-cache"); err == nil {
		cmd := exec.Command("gtk-update-icon-cache", "-f", "-t", themeDir)
		if err := cmd.Run(); err != nil {
			log.Debug("gtk-update-icon-cache failed: %v", err)
		}
	}

	// KDE Plasma
	for _, kbuildsycoca := range []string{"kbuildsycoca6", "kbuildsycoca5"} {
		if _, err := exec.LookPath(kbuildsycoca); err == nil {
			cmd := exec.Command(kbuildsycoca)
			if err := cmd.Run(); err != nil {
				log.Debug("%s failed: %v", kbuildsycoca, err)
			}
			break
		}
	}
}

func (m *Manager) copyNixStoreDesktop(binary string) ([]string, error) {
	currentAppsDir := filepath.Join(m.CurrentLink(), "share", "applications")

	entries, err := os.ReadDir(currentAppsDir)
	if err != nil {
		return nil, fmt.Errorf("reading nix store applications: %w", err)
	}

	var created []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) != ".desktop" {
			continue
		}

		srcPath := filepath.Join(currentAppsDir, name)
		dstPath := filepath.Join(m.ApplicationsDir(), name)

		virtualTarget, err := os.Readlink(srcPath)
		if err != nil {
			return created, fmt.Errorf("reading symlink %s: %w", srcPath, err)
		}

		realPath := m.virtualToRealStorePath(virtualTarget)

		content, err := os.ReadFile(realPath)
		if err != nil {
			return created, fmt.Errorf("reading desktop file %s: %w", realPath, err)
		}

		content = rewriteDesktopExecLines(content, m.BinDir(), binary)

		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			if os.IsExist(err) {
				continue
			}
			return created, fmt.Errorf("writing desktop file %s: %w", dstPath, err)
		}

		created = append(created, dstPath)
		log.Debug("Copied desktop file: %s (from %s)", dstPath, realPath)
	}

	return created, nil
}

func (m *Manager) virtualToRealStorePath(virtualPath string) string {
	const nixStorePrefix = "/nix/store/"
	if !strings.HasPrefix(virtualPath, nixStorePrefix) {
		return virtualPath
	}
	hashAndName := strings.TrimPrefix(virtualPath, nixStorePrefix)
	return filepath.Join(m.nixPortableLocation, ".nix-portable", "nix", "store", hashAndName)
}

var execLineRegex = regexp.MustCompile(`(?m)^Exec=/nix/store/[^/]+/bin/([^\s]+)(.*)$`)

func rewriteDesktopExecLines(content []byte, binDir, binary string) []byte {
	return execLineRegex.ReplaceAll(content, []byte("Exec="+binDir+"/"+binary+"$2"))
}

type desktopTemplateData struct {
	BinaryName    string
	Name          string
	GenericName   string
	CategoriesStr string
	BinDir        string
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

func (m *Manager) writeEmbeddedIcon(binary string) (string, error) {
	iconData, err := embeddedIcons.ReadFile("icons/" + binary + ".svg")
	if err != nil {
		return "", fmt.Errorf("reading embedded icon: %w", err)
	}

	iconPath := filepath.Join(m.IconsDir(), binary+".svg")
	if err := os.WriteFile(iconPath, iconData, 0644); err != nil {
		return "", fmt.Errorf("writing icon: %w", err)
	}

	log.Debug("Wrote icon: %s", iconPath)
	return iconPath, nil
}
