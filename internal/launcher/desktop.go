package launcher

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
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
Exec={{.BinaryName}} %f
Icon={{.BinaryName}}
Categories={{.CategoriesStr}};
`

func (m *Manager) ShareDir() string {
	return filepath.Join(m.profileDir, "share")
}

func (m *Manager) ApplicationsDir() string {
	return filepath.Join(m.ShareDir(), "applications")
}

func (m *Manager) IconsDir() string {
	return filepath.Join(m.ShareDir(), "icons", "hicolor", "scalable", "apps")
}

func (m *Manager) GenerateDesktopFiles(entries []DesktopEntry) error {
	appsDir := m.ApplicationsDir()
	iconsDir := m.IconsDir()

	if err := os.RemoveAll(appsDir); err != nil {
		return fmt.Errorf("removing old applications directory: %w", err)
	}

	if err := os.MkdirAll(appsDir, 0755); err != nil {
		return fmt.Errorf("creating applications directory: %w", err)
	}

	if err := os.MkdirAll(iconsDir, 0755); err != nil {
		return fmt.Errorf("creating icons directory: %w", err)
	}

	tmpl, err := template.New("desktop").Parse(desktopTemplate)
	if err != nil {
		return fmt.Errorf("parsing desktop template: %w", err)
	}

	for _, entry := range entries {
		switch e := entry.(type) {
		case NixStoreDesktop:
			if err := m.copyNixStoreDesktop(e.BinaryName); err != nil {
				log.Debug("Failed to copy desktop file for %s: %v", e.BinaryName, err)
			}
		case GeneratedDesktop:
			if err := m.generateDesktopFile(tmpl, e); err != nil {
				return fmt.Errorf("generating desktop file for %s: %w", e.BinaryName, err)
			}
			if err := m.writeEmbeddedIcon(e.BinaryName); err != nil {
				log.Debug("No embedded icon for %s: %v", e.BinaryName, err)
			}
		}
	}

	return nil
}

func (m *Manager) copyNixStoreDesktop(binary string) error {
	currentAppsDir := filepath.Join(m.CurrentLink(), "share", "applications")

	entries, err := os.ReadDir(currentAppsDir)
	if err != nil {
		return fmt.Errorf("reading nix store applications: %w", err)
	}

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
			return fmt.Errorf("reading symlink %s: %w", srcPath, err)
		}

		realPath := m.virtualToRealStorePath(virtualTarget)

		content, err := os.ReadFile(realPath)
		if err != nil {
			return fmt.Errorf("reading desktop file %s: %w", realPath, err)
		}

		content = rewriteDesktopExecLines(content, binary)

		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			if os.IsExist(err) {
				continue
			}
			return fmt.Errorf("writing desktop file %s: %w", dstPath, err)
		}

		log.Debug("Copied desktop file: %s (from %s)", dstPath, realPath)
	}

	return nil
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

func rewriteDesktopExecLines(content []byte, binary string) []byte {
	return execLineRegex.ReplaceAll(content, []byte("Exec="+binary+"$2"))
}

func (m *Manager) generateDesktopFile(tmpl *template.Template, entry GeneratedDesktop) error {
	desktopPath := filepath.Join(m.ApplicationsDir(), entry.BinaryName+".desktop")

	f, err := os.Create(desktopPath)
	if err != nil {
		return fmt.Errorf("creating desktop file: %w", err)
	}

	execErr := tmpl.Execute(f, entry)
	closeErr := f.Close()

	if execErr != nil {
		return fmt.Errorf("writing desktop file: %w", execErr)
	}
	if closeErr != nil {
		return fmt.Errorf("closing desktop file: %w", closeErr)
	}

	log.Debug("Generated desktop file: %s", desktopPath)
	return nil
}

func (m *Manager) writeEmbeddedIcon(binary string) error {
	iconData, err := embeddedIcons.ReadFile("icons/" + binary + ".svg")
	if err != nil {
		return fmt.Errorf("reading embedded icon: %w", err)
	}

	iconPath := filepath.Join(m.IconsDir(), binary+".svg")
	if err := os.WriteFile(iconPath, iconData, 0644); err != nil {
		return fmt.Errorf("writing icon: %w", err)
	}

	log.Debug("Wrote icon: %s", iconPath)
	return nil
}
